package main

import (
	"fmt"
	"net"
	"os"

	"github.com/KindCloud97/transactionapi/etherscan"
	"github.com/KindCloud97/transactionapi/gen/proto/transactionapi"
	"github.com/KindCloud97/transactionapi/server"
	"github.com/KindCloud97/transactionapi/store"
	"github.com/KindCloud97/transactionapi/updater"

	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Err(err).Msg("load environment variable")
	}
	apiKey := os.Getenv("APIKEY")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	port := os.Getenv("PORT")

	uri := "mongodb+srv://" + username + ":" + password +
		"@cluster0.imbqypd.mongodb.net/?retryWrites=true&w=majority"

	mongo, err := store.ConnectMongo(uri)
	if err != nil {
		log.Fatal().Err(err).Msg("error connect to mongoDB")
	}

	ethClient := etherscan.New(apiKey)

	u := updater.New(ethClient, mongo)

	go func() {
		err = u.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("updater failed")
		}
	}()

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("listen failed")
	}

	grpcServer := grpc.NewServer()
	log.Info().Str("addr", addr).Msg("start server")
	transactionapi.RegisterTransactionServiceServer(grpcServer, server.New(mongo))
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
