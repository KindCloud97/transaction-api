package main

import (
	"context"
	"fmt"

	"github.com/KindCloud97/transactionapi/gen/proto/transactionapi"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	conn, err := grpc.Dial("transactionapi.fly.dev:443", grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	if err != nil {
		log.Fatal().Err(err).Msg("connection failed")
	}

	client := transactionapi.NewTransactionServiceClient(conn)
	resp, err := client.GetTransactions(context.Background(), &transactionapi.GetTransactionsRequest{
		Id:        "",
		To:        "",
		BlockId:   0,
		Timestamp: "",
		Value:     "",
		Gas:       "",
		Page: &transactionapi.PageRequest{
			Size: 3,
			Num:  1,
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("get transactions failed")
	}
	fmt.Println(protojson.Format(resp))

}
