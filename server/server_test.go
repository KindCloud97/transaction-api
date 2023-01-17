package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/KindCloud97/transactionapi/gen/proto/transactionapi"
	"github.com/KindCloud97/transactionapi/store"
	storemock "github.com/KindCloud97/transactionapi/store/mocks"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestServer_GetTransactions(t *testing.T) {
	type fields struct {
		mongo *store.Mongo
	}
	type args struct {
		ctx context.Context
		r   *transactionapi.GetTransactionsRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		setupMock func(m *storemock.Storer)
		want      *transactionapi.GetTransactionsResponse
		wantErr   bool
	}{
		{
			name: "success",
			fields: fields{
				mongo: &store.Mongo{},
			},
			args: args{
				ctx: context.Background(),
				r: &transactionapi.GetTransactionsRequest{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "testGas",
					Page:      &transactionapi.PageRequest{Size: 1, Num: 1},
				},
			},
			setupMock: func(m *storemock.Storer) {
				m.EXPECT().FindPage(store.Transaction{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "testGas",
				}, 1, 1).Return([]store.Transaction{
					{
						Id:        "testId",
						From:      "Jotaro",
						BlockId:   10,
						To:        "Jolyne",
						Timestamp: "testTime",
						Value:     "10000",
						Gas:       "testGas",
					},
				}, nil).Once()
			},
			want: &transactionapi.GetTransactionsResponse{
				Transactions: []*transactionapi.Transaction{
					{
						Id:        "testId",
						From:      "Jotaro",
						BlockId:   10,
						To:        "Jolyne",
						Timestamp: "testTime",
						Value:     "0.000000000000065536",
						Gas:       "testGas",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "not_found",
			fields: fields{
				mongo: &store.Mongo{},
			},
			args: args{
				ctx: context.Background(),
				r: &transactionapi.GetTransactionsRequest{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "testGas",
					Page:      &transactionapi.PageRequest{Size: 1, Num: 1},
				},
			},
			setupMock: func(m *storemock.Storer) {
				m.EXPECT().FindPage(store.Transaction{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "testGas",
				}, 1, 1).Return(nil, status.Error(codes.NotFound, "page not found")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "internal_error",
			fields: fields{
				mongo: &store.Mongo{},
			},
			args: args{
				ctx: context.Background(),
				r: &transactionapi.GetTransactionsRequest{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "",
					Gas:       "testGas",
					Page:      &transactionapi.PageRequest{Size: 1, Num: 1},
				},
			},
			setupMock: func(m *storemock.Storer) {
				m.EXPECT().FindPage(store.Transaction{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "",
					Gas:       "testGas",
				}, 1, 1).Return([]store.Transaction{
					{
						Id:        "testId",
						From:      "Jotaro",
						BlockId:   10,
						To:        "Jolyne",
						Timestamp: "testTime",
						Value:     "",
						Gas:       "testGas",
					},
				}, nil).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error_page_number",
			fields: fields{
				mongo: &store.Mongo{},
			},
			args: args{
				ctx: context.Background(),
				r: &transactionapi.GetTransactionsRequest{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "testGas",
					Page:      &transactionapi.PageRequest{Size: 1, Num: 0},
				},
			},
			setupMock: func(m *storemock.Storer) {

			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error_page_size",
			fields: fields{
				mongo: &store.Mongo{},
			},
			args: args{
				ctx: context.Background(),
				r: &transactionapi.GetTransactionsRequest{
					Id:        "testId",
					From:      "Jotaro",
					BlockId:   10,
					To:        "Jolyne",
					Timestamp: "testTime",
					Value:     "10000",
					Gas:       "testGas",
					Page:      &transactionapi.PageRequest{Size: 0, Num: 1},
				},
			},
			setupMock: func(m *storemock.Storer) {

			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := storemock.NewStorer(t)
			s := &Server{
				storer: m,
			}

			if tt.setupMock != nil {
				tt.setupMock(m)
			}
			got, err := s.GetTransactions(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.GetTransactions() = %v, want %v\nDiff:\n%s", got, tt.want,
					cmp.Diff(got, tt.want, protocmp.Transform()))
			}
		})
	}
}
