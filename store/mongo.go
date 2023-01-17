package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockery --name Storer --with-expecter
type Storer interface {
	FindPage(tran Transaction, pageNum, pageSize int) ([]Transaction, error)
	GetLastBlock() (int64, error)
	InsertOne(ctx context.Context, t Transaction) error
}

type Transaction struct {
	Id        string `bson:"id,omitempty"`
	From      string `bson:"from,omitempty"`
	BlockId   int64  `bson:"block_id,omitempty"`
	To        string `bson:"to,omitempty"`
	Timestamp string `bson:"timestamp,omitempty"`
	Value     string `bson:"value,omitempty"`
	Gas       string `bson:"gas,omitempty"`
}

type Search struct {
	Trans Transaction
	Total int64
}

type Mongo struct {
	col   *mongo.Collection
	pages map[uint64]Search
	mu    sync.Mutex
}

type mongoPaginate struct {
	limit int64
	page  int64
}

func newMongoPaginate(limit, page int) *mongoPaginate {
	return &mongoPaginate{
		limit: int64(limit),
		page:  int64(page),
	}
}

func ConnectMongo(uri string) (*Mongo, error) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	collection := client.Database("baz").Collection("qux")

	return &Mongo{
		col:   collection,
		pages: make(map[uint64]Search),
	}, nil
}

func (m *Mongo) FindPage(tran Transaction, pageNum, pageSize int) ([]Transaction, error) {
	opts := newMongoPaginate(pageSize, pageNum).getPaginatedOpts()

	cursor, err := m.col.Find(context.Background(), tran, opts)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("find: %w", err)
		}
		return nil, fmt.Errorf("find: %w", err)
	}
	defer cursor.Close(context.Background())

	hash, err := hashstructure.Hash(tran, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}

	total, err := m.col.CountDocuments(context.Background(), tran)
	if err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}

	m.mu.Lock()
	m.pages[hash] = Search{
		Trans: tran,
		Total: total,
	}
	m.mu.Unlock()

	if pageNum*pageSize > int(m.pages[hash].Total) {
		return nil, fmt.Errorf("out of bounds")
	}

	var results []Transaction
	for i := 0; i < pageSize; i++ {
		if cursor.TryNext(context.Background()) {
			var t Transaction
			if err := cursor.Decode(&t); err != nil {
				return nil, fmt.Errorf("decode: %w", err)
			}

			results = append(results, t)

			continue
		}

		if err := cursor.Err(); err != nil {
			return nil, fmt.Errorf("cursor: %w", err)
		}

		if cursor.ID() == 0 {
			break
		}
	}

	return results, nil
}

func (m *Mongo) GetLastBlock() (int64, error) {
	opts := options.FindOne().SetSort(bson.D{{
		Key:   "block_id",
		Value: -1,
	}})
	var result Transaction
	err := m.col.FindOne(context.Background(), bson.D{}, opts).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("find one: %w", err)
		}
		return 0, fmt.Errorf("find one: %w", err)
	}

	return result.BlockId, nil
}

func (m *Mongo) InsertOne(ctx context.Context, t Transaction) error {
	_, err := m.col.InsertOne(context.Background(), t)
	if err != nil {
		return fmt.Errorf("insert one: %w", err)
	}
	return nil
}

func (mp *mongoPaginate) getPaginatedOpts() *options.FindOptions {
	l := mp.limit
	skip := mp.page*mp.limit - mp.limit
	fOpt := options.FindOptions{Limit: &l, Skip: &skip}

	return &fOpt
}
