package updater

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KindCloud97/transactionapi/etherscan"
	"github.com/KindCloud97/transactionapi/store"
	"github.com/rs/zerolog/log"
)

const numOldBlocks = 100

type Updater struct {
	client etherscan.Client
	store  *store.Mongo
}

func New(client etherscan.Client, store *store.Mongo) *Updater {
	return &Updater{
		client: client,
		store:  store,
	}
}

func (s *Updater) loadMissingTransactions(lastBlockDB int64,
	latestBlockBC int64) error {
	if latestBlockBC-lastBlockDB > numOldBlocks {
		lastBlockDB = latestBlockBC - numOldBlocks
	}

	for i := latestBlockBC; i > lastBlockDB; i-- {
		err := s.fetchAndSaveTransactionsInBlock(i)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Updater) fetchAndSaveTransactionsInBlock(blockNum int64) error {
	log.Debug().Int64("block id", blockNum).Msg("fetching and saving all transactions in the block")

	block, err := s.client.GetBlock(blockNum)
	if err != nil {
		return err
	}

	for _, trId := range block.Transactions {
		transaction, err := s.client.GetTransaction(trId)
		if err != nil {
			return fmt.Errorf("get transaction ID = %s error: %w", trId, err)
		}

		t := apiTransToDBTrans(block, transaction)
		err = s.store.InsertOne(context.Background(), t)
		if err != nil {
			return fmt.Errorf("api transaction to db transaction: %w", err)
		}

		log.Debug().Str("transaction id:", trId).Send()
	}

	return nil
}

func (s *Updater) Start() error {
	for {
		latestBlockBC, err := s.client.GetLatestBlock()
		if err != nil {
			return fmt.Errorf("get latest block: %w", err)
		}
		lastBlockDB, err := s.store.GetLastBlock()
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "no documents in result"):
				lastBlockDB = -1
			default:
				return fmt.Errorf("get last block: %w", err)
			}
		}

		err = s.loadMissingTransactions(lastBlockDB, latestBlockBC)
		if err != nil {
			return fmt.Errorf("load missing transactions: %w", err)
		}
		time.Sleep(1 * time.Minute)
	}
}

func apiTransToDBTrans(
	b etherscan.Block,
	t etherscan.Transaction) store.Transaction {
	return store.Transaction{
		Id:        t.Hash,
		To:        t.To,
		BlockId:   b.Number,
		Timestamp: b.Timestamp,
		Value:     t.Value,
		Gas:       t.GasPrice,
	}
}
