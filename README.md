# Transaction-API

API for getting a list of transactions in the Ethereum blockchain

## Usage
This package implements protobuf service that you can see below.
```protobuf
service TransactionService {
  rpc GetTransactions(GetTransactionsRequest) returns (GetTransactionsResponse) {}
}
```
Use `GetTransactionsRequest` object to filtering through all transactions in database. 
```protobuf
message GetTransactionsRequest {
  string id = 1;
  string from = 2;
  int64 block_id = 3;
  string to = 4;
  string timestamp = 5;
  string value = 6;
  string gas = 7;
  PageRequest page = 8;
}
```
The API supports paging, to use it you can pass `PageRequest` with `size`(page size) and `num`(page number).
```protobuf
message PageRequest {
  int64 size = 2;
  int64 num = 3;
}
```
The method `GetTransactionsResponse` returns slice of `Transaction`.
```protobuf
message GetTransactionsResponse {
  repeated Transaction transactions = 1;
}
```
`Transaction` includes the following fields.
```protobuf
message Transaction {
  string id = 1;
  string from = 2;
  int64 block_id = 3;
  string to = 4;
  string timestamp = 5;
  string value = 6;
  string gas = 7;
}
```

##### The API deployed on Fly.io you can try it on <https://transactionapi.fly.dev:443>
