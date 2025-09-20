package models

import (
	"time"
)

type EquityPriceData struct {
	Timestamps    time.Time
	Open          float32
	Close         float32
	High          float32
	Low           float32
	Volume        float32
	PercentChange float32
}

func (e EquityPriceData) GetTime() time.Time {
	return e.Timestamps
}

type Ticker struct {
	Symbol     string
	Low52Week  float32
	High52Week float32
}
