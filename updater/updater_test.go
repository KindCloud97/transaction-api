package updater

import (
	"context"
	"errors"
	"testing"

	"github.com/KindCloud97/transactionapi/etherscan"
	"github.com/KindCloud97/transactionapi/store"
	storemock "github.com/KindCloud97/transactionapi/store/mocks"
	updatermock "github.com/KindCloud97/transactionapi/updater/mocks"
	"github.com/google/go-cmp/cmp"
)

func Test_loadTransactions(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(ms *storemock.Storer, mg *updatermock.EtherScanner)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(ms *storemock.Storer, mu *updatermock.EtherScanner) {
				mu.EXPECT().GetLatestBlock().Return(10, nil).Once()
				ms.EXPECT().GetLastBlock().Return(9, nil).Once()

				mu.EXPECT().GetBlock(int64(10)).Return(etherscan.Block{
					Number:       10,
					Timestamp:    "testTime",
					Transactions: []string{"testHash"},
				}, nil).Once()
				mu.EXPECT().GetTransaction("testHash").Return(etherscan.Transaction{
					Hash:     "testHash",
					From:     "test1",
					To:       "test2",
					Value:    "10000",
					GasPrice: "100",
				}, nil).Once()
				ms.EXPECT().InsertOne(context.Background(), store.Transaction{
					Id:        "testHash",
					From:      "test1",
					BlockId:   10,
					To:        "test2",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "100",
				}).Return(nil)

			},
			wantErr: false,
		},
		{
			name: "error_get_latest_block_BC",
			setupMock: func(ms *storemock.Storer, mu *updatermock.EtherScanner) {
				mu.EXPECT().GetLatestBlock().Return(0, errors.New("smth went wrong")).Once()
			},
			wantErr: true,
		},
		{
			name: "error_get_last_block_DB",
			setupMock: func(ms *storemock.Storer, mu *updatermock.EtherScanner) {
				mu.EXPECT().GetLatestBlock().Return(10, nil).Once()
				ms.EXPECT().GetLastBlock().Return(0, errors.New("smth went wrong")).Once()
			},
			wantErr: true,
		},
		{
			name: "error_get_block",
			setupMock: func(ms *storemock.Storer, mu *updatermock.EtherScanner) {
				mu.EXPECT().GetLatestBlock().Return(10, nil).Once()
				ms.EXPECT().GetLastBlock().Return(9, nil).Once()

				mu.EXPECT().GetBlock(int64(10)).
					Return(etherscan.Block{}, errors.New("smth went wrong")).Once()
			},
			wantErr: true,
		},
		{
			name: "error_get_transaction",
			setupMock: func(ms *storemock.Storer, mu *updatermock.EtherScanner) {
				mu.EXPECT().GetLatestBlock().Return(10, nil).Once()
				ms.EXPECT().GetLastBlock().Return(9, nil).Once()

				mu.EXPECT().GetBlock(int64(10)).Return(etherscan.Block{
					Number:       10,
					Timestamp:    "testTime",
					Transactions: []string{"testHash"},
				}, nil).Once()
				mu.EXPECT().GetTransaction("testHash").
					Return(etherscan.Transaction{}, errors.New("smth went wrong")).Once()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := storemock.NewStorer(t)
			mu := updatermock.NewEtherScanner(t)
			u := &Updater{
				escan:  mu,
				storer: ms,
			}

			if tt.setupMock != nil {
				tt.setupMock(ms, mu)
			}

			if err := loadTransactions(u); (err != nil) != tt.wantErr {
				t.Errorf("loadTransactions() error = %v, wantErr %v\nDiff:\n%s", err, tt.wantErr,
					cmp.Diff(err, tt.wantErr))
			}
		})
	}
}
