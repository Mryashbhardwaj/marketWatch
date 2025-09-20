package models

import (
	"time"
)

type MFPriceData struct {
	Timestamps    time.Time
	Price         float32
	PercentChange float32
}

func (m MFPriceData) GetTime() time.Time {
	return m.Timestamps
}
func (m MFPriceData) GetPrice() float64 {
	return float64(m.Price)
}

type MFSummary struct {
	Name                            string
	ISIN                            string
	HoldingSince                    time.Duration
	LastInvestment                  time.Duration
	HoldingFrom                     time.Duration
	CurrentValue                    float64
	InvestedValue                   float64
	AllTimeAbsoluteReturn           float64
	AllTimeAbsoluteReturnPercentage float64
	XIRR                            float64
	CAGR                            float64
}

type MFHoldingsData struct {
	Timestamps     time.Time
	TotalValue     float64
	TotalUnitsHeld float64
	Transaction    float64
}

func (m MFHoldingsData) GetTime() time.Time {
	return m.Timestamps
}
