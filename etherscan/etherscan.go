package etherscan

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/ratelimit"
)

//go:generate mockery --name HttpGetter --with-expecter
type HttpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

type Client struct {
	apiKey     string
	rateLimit  ratelimit.Limiter
	httpGetter HttpGetter
}

type block struct {
	Result struct {
		Number       string   `bson:"number,omitempty"`
		Timestamp    string   `bson:"timestamp,omitempty"`
		Transactions []string `bson:"transactions,omitempty"`
	} `bson:"result,omitempty"`
}

type Block struct {
	Number       int64
	Timestamp    string
	Transactions []string
}

type transaction struct {
	Result struct {
		Hash     string `bson:"hash,omitempty"`
		From     string `bson:"from,omitempty"`
		To       string `bson:"to,omitempty"`
		Value    string `bson:"value,omitempty"`
		GasPrice string `bson:"gasPrice,omitempty"`
	} `bson:"result,omitempty"`
}

type Transaction struct {
	Hash     string
	From     string
	To       string
	Value    string
	GasPrice string
}

type Response struct {
	Id     int    `bson:"id,omitempty"`
	Result string `bson:"result,omitempty"`
}

type Error struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Result  string `json:"result,omitempty"`
}

func New(apikey string, httpGetter HttpGetter) *Client {
	return &Client{
		apiKey:     apikey,
		rateLimit:  ratelimit.New(4, ratelimit.WithoutSlack),
		httpGetter: httpGetter,
	}
}

func (err *Error) Error() string {
	return err.Result
}

func (c *Client) GetLatestBlock() (int64, error) {
	c.rateLimit.Take()

	resp, err := c.httpGetter.Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=" + c.apiKey)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		var respErr Error
		err := json.Unmarshal(body, &respErr)
		if err != nil {
			return 0, err
		}
		return 0, &respErr
	}

	var r Response
	err = json.Unmarshal(body, &r)
	if err != nil {
		return 0, err
	}

	hex := strings.TrimPrefix(r.Result, "0x")
	blockNum, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return 0, err
	}

	return blockNum, nil
}

func (c *Client) GetBlock(blockNum int64) (Block, error) {
	c.rateLimit.Take()

	resp, err := c.httpGetter.Get("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=" +
		fmt.Sprintf("0x%x", blockNum) + "&boolean=false&apikey=" + c.apiKey)
	if err != nil {
		return Block{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Block{}, err
	}

	if resp.StatusCode != 200 {
		var respErr Error
		err := json.Unmarshal(body, &respErr)
		if err != nil {
			return Block{}, err
		}
		return Block{}, &respErr
	}

	var block block
	err = json.Unmarshal(body, &block)
	if err != nil {
		return Block{}, err
	}

	return block.toBlock()
}

func (b *block) toBlock() (Block, error) {
	hex := strings.TrimPrefix(b.Result.Number, "0x")
	num, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return Block{}, err
	}

	return Block{
		Number:       num,
		Timestamp:    b.Result.Timestamp,
		Transactions: b.Result.Transactions,
	}, nil
}

func (c *Client) GetTransaction(hash string) (Transaction, error) {
	c.rateLimit.Take()

	resp, err := c.httpGetter.Get("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=" +
		hash + "&apikey=" + c.apiKey)
	if err != nil {
		return Transaction{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Transaction{}, err
	}

	if resp.StatusCode != 200 {
		var respErr Error
		err := json.Unmarshal(body, &respErr)
		if err != nil {
			return Transaction{}, err
		}
		return Transaction{}, &respErr
	}

	var t transaction
	err = json.Unmarshal(body, &t)
	if err != nil {
		return Transaction{}, err
	}

	return t.toTransaction()
}

func (t *transaction) toTransaction() (Transaction, error) {
	hex := strings.TrimPrefix(t.Result.Value, "0x")
	value := new(big.Int)

	value, ok := value.SetString(hex, 16)
	if !ok {
		return Transaction{}, fmt.Errorf("set string error")
	}

	return Transaction{
		Hash:     t.Result.Hash,
		From:     t.Result.From,
		To:       t.Result.To,
		Value:    value.String(),
		GasPrice: t.Result.GasPrice,
	}, nil
}
