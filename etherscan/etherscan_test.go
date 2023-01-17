package etherscan_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	_ "embed"

	"github.com/KindCloud97/transactionapi/etherscan"
	etherscanmock "github.com/KindCloud97/transactionapi/etherscan/mocks"
	"github.com/google/go-cmp/cmp"
)

var (
	//go:embed test_data/resp_get_latest_block.json
	respGetLatestBlock string
	//go:embed test_data/resp_get_block.json
	respGetBlock string
	//go:embed test_data/resp_get_transaction.json
	respGetTransaction string
)

func TestClient_GetLatestBlock(t *testing.T) {
	type fields struct {
		apiKey string
	}
	tests := []struct {
		name      string
		fields    fields
		setupMock func(m *etherscanmock.HttpGetter)
		want      int64
		wantErr   bool
	}{
		{
			name: "success",
			fields: fields{
				apiKey: "apiKey",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=apiKey").Return(&http.Response{
					Status:     "OK",
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(respGetLatestBlock)),
				}, nil).Once()
			},
			want:    16426811,
			wantErr: false,
		},
		{
			name: "error_get",
			fields: fields{
				apiKey: "apiKey",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=apiKey").Return(nil, errors.New("smth went wrong")).Once()
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "error_wrong_api_key",
			fields: fields{
				apiKey: "",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=").Return(&http.Response{
					Status:     "NOTOK",
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK","result":"Invalid API Key"}`)),
				}, nil).Once()
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "error_invalid_action",
			fields: fields{
				apiKey: "apiKey",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=apiKey").Return(&http.Response{
					Status:     "NOTOK",
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK-Missing/Invalid API Key, rate limit of 1/5sec applied","result":"Error! Missing Or invalid Action name"}`)),
				}, nil).Once()
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := etherscanmock.NewHttpGetter(t)
			c := etherscan.New(tt.fields.apiKey, m)

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			got, err := c.GetLatestBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetLatestBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.GetLatestBlock() = %v, want %v \nDiff:\n%s", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestClient_GetBlock(t *testing.T) {
	type fields struct {
		apiKey string
	}
	type args struct {
		blockNum int64
	}
	tests := []struct {
		name      string
		fields    fields
		setupMock func(m *etherscanmock.HttpGetter)
		args      args
		want      etherscan.Block
		wantErr   bool
	}{
		{
			name: "success",
			fields: fields{
				apiKey: "apiKey",
			},
			args: args{
				blockNum: 12989009,
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag="+
					fmt.Sprintf("0x%x", 12989009)+"&boolean=false&apikey=apiKey").Return(&http.Response{
					Status:     "OK",
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(respGetBlock)),
				}, nil).Once()
			},
			want: etherscan.Block{
				Number:    12989009,
				Timestamp: "0x6110bab2",
				Transactions: []string{"0x40330c87750aa1ba1908a787b9a42d0828e53d73100ef61ae8a4d925329587b5",
					"0x6fa2208790f1154b81fc805dd7565679d8a8cc26112812ba1767e1af44c35dd4",
					"0xe31d8a1f28d4ba5a794e877d65f83032e3393809686f53fa805383ab5c2d3a3c",
					"0xa6a83df3ca7b01c5138ec05be48ff52c7293ba60c839daa55613f6f1c41fdace",
					"0x4e46edeb68a62dde4ed081fae5efffc1fb5f84957b5b3b558cdf2aa5c2621e17",
					"0x356ee444241ae2bb4ce9f77cdbf98cda9ffd6da244217f55465716300c425e82",
					"0x1a4ec2019a3f8b1934069fceff431e1370dcc13f7b2561fe0550cc50ab5f4bbc",
					"0xad7994bc966aed17be5d0b6252babef3f56e0b3f35833e9ac414b45ed80dac93"},
			},
			wantErr: false,
		},
		{
			name: "error_get",
			fields: fields{
				apiKey: "apiKey",
			},
			args: args{
				blockNum: 12989009,
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag="+
					fmt.Sprintf("0x%x", 12989009)+"&boolean=false&apikey=apiKey").Return(nil, errors.New("smth went wrong")).Once()
			},
			want:    etherscan.Block{},
			wantErr: true,
		},
		{
			name: "error_wrong_api_key",
			fields: fields{
				apiKey: "",
			},
			args: args{
				blockNum: 12989009,
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag="+
					fmt.Sprintf("0x%x", 12989009)+"&boolean=false&apikey=").Return(&http.Response{
					Status:     "NOTOK",
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK","result":"Invalid API Key"}`)),
				}, nil).Once()
			},
			want:    etherscan.Block{},
			wantErr: true,
		},
		{
			name: "error_wrong_block_number",
			fields: fields{
				apiKey: "apiKey",
			},
			args: args{
				blockNum: 0,
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag="+
					fmt.Sprintf("0x%x", 0)+"&boolean=false&apikey=apiKey").Return(&http.Response{
					Status:     "NOTOK",
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK","result":"Invalid API Key"}`)),
				}, nil).Once()
			},
			want:    etherscan.Block{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := etherscanmock.NewHttpGetter(t)
			c := etherscan.New(tt.fields.apiKey, m)

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			got, err := c.GetBlock(tt.args.blockNum)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetBlock() = %v, want %v \nDiff:\n%s", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestClient_GetTransaction(t *testing.T) {
	type fields struct {
		apiKey string
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		setupMock func(m *etherscanmock.HttpGetter)
		want      etherscan.Transaction
		wantErr   bool
	}{
		{
			name: "success",
			fields: fields{
				apiKey: "apiKey",
			},
			args: args{
				hash: "0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash="+
					"0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44"+"&apikey=apiKey").Return(&http.Response{
					Status:     "OK",
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(respGetTransaction)),
				}, nil).Once()
			},
			want: etherscan.Transaction{
				Hash:     "0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44",
				From:     "0x00192fb10df37c9fb26829eb2cc623cd1bf599e8",
				To:       "0xc67f4e626ee4d3f272c2fb31bad60761ab55ed9f",
				Value:    "7165918000000000",
				GasPrice: "0x19f017ef49",
			},
			wantErr: false,
		},
		{
			name: "error_get",
			fields: fields{
				apiKey: "apiKey",
			},
			args: args{
				hash: "0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash="+
					"0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44"+"&apikey=apiKey").Return(nil, errors.New("smth went wrong")).Once()
			},
			want:    etherscan.Transaction{},
			wantErr: true,
		},
		{
			name: "error_wrong_api_key",
			fields: fields{
				apiKey: "",
			},
			args: args{
				hash: "0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash="+
					"0xbc78ab8a9e9a0bca7d0321a27b2c03addeae08ba81ea98b03cd3dd237eabed44"+"&apikey=").Return(&http.Response{
					Status:     "NOTOK",
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK","result":"Invalid API Key"}`)),
				}, nil).Once()
			},
			want:    etherscan.Transaction{},
			wantErr: true,
		},
		{
			name: "error_wrong_hash",
			fields: fields{
				apiKey: "apiKey",
			},
			args: args{
				hash: "",
			},
			setupMock: func(m *etherscanmock.HttpGetter) {
				m.EXPECT().Get("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash="+
					""+"&apikey=apiKey").Return(&http.Response{
					Status:     "NOTOK",
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK","result":"Invalid API Key"}`)),
				}, nil).Once()
			},
			want:    etherscan.Transaction{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := etherscanmock.NewHttpGetter(t)
			c := etherscan.New(tt.fields.apiKey, m)

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			got, err := c.GetTransaction(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetTransaction() = %v, want %v\nDiff:\n%s", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
