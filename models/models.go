package models

import "sync"

type UserID string

type MarketID string
type Price int

type UserBalance struct {
	Balance int
	Locked  int
}

type Market struct {
	Book *Orderbook
}

type StockBalances struct {
	sync.Mutex
	StockBalances map[UserID]interface{}
}

type Orderbook struct {
	sync.Mutex
	YesOrders map[Price]interface{}
	NoOrders  map[Price]interface{}
}

type InrBalances struct {
	sync.Mutex
	InrBalances map[UserID]UserBalance
}
