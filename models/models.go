package models

import "sync"

type UserID string
type StockID string
type MarketID string

type UserBalance struct {
	Balance int
	Locked  int
}

type StockBalances struct {
	Mu       sync.Mutex
	Balances map[StockID]interface{}
}

type Orderbook struct {
	Mu     sync.Mutex
	Orders map[MarketID]interface{}
}

type InrBalances struct {
	Mu       sync.Mutex
	Balances map[UserID]UserBalance
}
