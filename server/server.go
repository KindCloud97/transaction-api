package server

import (
	"context"
	"fmt"
	"github.com/KindCloud97/transactionapi/gen/proto/transactionapi"
	"github.com/KindCloud97/transactionapi/store"
	"math/big"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ transactionapi.TransactionServiceServer = (*Server)(nil)

type Server struct {
	transactionapi.UnimplementedTransactionServiceServer
	mongo *store.Mongo
}

func New(mongo *store.Mongo) *Server {
	return &Server{
		mongo: mongo,
	}
}

func (s *Server) GetTransactions(ctx context.Context,
	r *transactionapi.GetTransactionsRequest) (*transactionapi.GetTransactionsResponse, error) {

	page := r.GetPage()

	if page.Num <= 0 {
		log.Error().Msg("error incorrect page number")
		return nil, status.Error(codes.InvalidArgument, "incorrect page number")
	}
	if page.Size <= 0 {
		log.Error().Msg("error incorrect page size")
		return nil, status.Error(codes.InvalidArgument, "incorrect page size")
	}

	trans, err := s.mongo.FindPage(store.Transaction{
		Id:        r.Id,
		To:        r.To,
		BlockId:   r.BlockId,
		Timestamp: r.Timestamp,
		Value:     r.Value,
		Gas:       r.Gas,
	}, int(page.Num), int(page.Size))
	if err != nil {
		log.Err(err).Msg("page not found")
		return nil, status.Error(codes.NotFound, "page not found")
	}

	converted, err := convertTransactions(trans)
	if err != nil {
		log.Err(err).Msg("convert transactions")
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &transactionapi.GetTransactionsResponse{
		Transactions: converted,
	}, nil
}

func convertTransactions(pages []store.Transaction) ([]*transactionapi.Transaction, error) {
	var transactions []*transactionapi.Transaction
	for _, trans := range pages {
		value, err := weiToEther(trans.Value)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, &transactionapi.Transaction{
			Id:        trans.Id,
			To:        trans.To,
			BlockId:   trans.BlockId,
			Timestamp: trans.Timestamp,
			Value:     value,
			Gas:       trans.Gas,
		})
	}

	return transactions, nil
}

func weiToEther(wei string) (string, error) {
	amount := new(big.Int)
	amount, ok := amount.SetString(wei, 16)
	if !ok {
		return "", fmt.Errorf("set string error")
	}

	compact_amount := big.NewInt(0)
	reminder := big.NewInt(0)
	divisor := big.NewInt(1e18)
	compact_amount.QuoRem(amount, divisor, reminder)
	return fmt.Sprintf("%v.%018s", compact_amount.String(), reminder.String()), nil
}
