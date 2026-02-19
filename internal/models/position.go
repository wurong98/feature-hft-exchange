package models

import "time"

type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

type Position struct {
	APIKey         string       `json:"-"`
	Symbol         string       `json:"symbol"`
	Side           PositionSide `json:"side"`
	EntryPrice     float64      `json:"entryPrice,string"`
	Size           float64      `json:"size,string"`
	Leverage       int          `json:"leverage"`
	Margin         float64      `json:"margin,string"`
	UnrealizedPNL  float64      `json:"unrealizedPnl,string"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

// Balance 账户余额
type Balance struct {
	APIKey     string  `json:"-"`
	Available  float64 `json:"available,string"`
	Frozen     float64 `json:"frozen,string"`
	TotalPNL   float64 `json:"totalPnl,string"`
}
