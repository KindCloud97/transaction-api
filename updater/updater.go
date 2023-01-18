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

const numOldBlocks = 1000

//go:generate mockery --name EtherScanner --with-expecter
type EtherScanner interface {
	GetBlock(blockNum int64) (etherscan.Block, error)
	GetLatestBlock() (int64, error)
	GetTransaction(hash string) (etherscan.Transaction, error)
}

type Updater struct {
	escan  EtherScanner
	storer store.Storer
}

func New(escan EtherScanner, storer store.Storer) *Updater {
	return &Updater{
		escan:  escan,
		storer: storer,
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

	block, err := s.escan.GetBlock(blockNum)
	if err != nil {
		return err
	}

	for _, trId := range block.Transactions {
		transaction, err := s.escan.GetTransaction(trId)
		if err != nil {
			return fmt.Errorf("get transaction ID = %s error: %w", trId, err)
		}

		t := apiTransToDBTrans(block, transaction)
		err = s.storer.InsertOne(context.Background(), t)
		if err != nil {
			return fmt.Errorf("insert one: %w", err)
		}

		log.Debug().Str("transaction id:", trId).Send()
	}

	return nil
}

func (s *Updater) Start() error {
	for {
		if err := loadTransactions(s); err != nil {
			return err
		}
		time.Sleep(1 * time.Minute)
	}
}

func loadTransactions(s *Updater) error {
	latestBlockBC, err := s.escan.GetLatestBlock()
	if err != nil {
		return fmt.Errorf("get latest block: %w", err)
	}

	lastBlockDB, err := s.storer.GetLastBlock()
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

	return nil
}

func apiTransToDBTrans(
	b etherscan.Block,
	t etherscan.Transaction) store.Transaction {
	return store.Transaction{
		Id:        t.Hash,
		From:      t.From,
		BlockId:   b.Number,
		To:        t.To,
		Timestamp: b.Timestamp,
		Value:     t.Value,
		Gas:       t.GasPrice,
	}
}
