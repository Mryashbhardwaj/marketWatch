package tradebook_service

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Mryashbhardwaj/marketAnalysis/models"
	"github.com/Mryashbhardwaj/marketAnalysis/utils"
)

func GetMutualFundsList() []string {
	var fundList []string
	for fundName, insi := range mutualFundsTradebook.AllFunds {
		fundList = append(fundList, fmt.Sprintf("%s:%s", fundName, insi))
	}
	return fundList
}

func GetEquityList() []ScriptName {
	return equityTradebook.AllScripts
}

func GetMFPriceTrendInTimeRange(symbol string, from, to time.Time) []models.EquityPriceData {
	if len(shareHistory[ScriptName(symbol)]) == 0 {
		return nil
	}
	startIndex := utils.MomentBinarySearch(shareHistory[ScriptName(symbol)], from)
	endIndex := utils.MomentBinarySearch(shareHistory[ScriptName(symbol)], to)

	requestedRange := shareHistory[ScriptName(symbol)][startIndex:endIndex]
	if len(requestedRange) == 0 {
		return nil
	}
	startPrice := requestedRange[0].Close
	for i, _ := range requestedRange {
		requestedRange[i].PercentChange = ((requestedRange[i].Close - startPrice) / startPrice) * 100
	}
	return requestedRange
}

func GetPriceTrendInTimeRange(symbol string, from, to time.Time) []models.EquityPriceData {
	if len(shareHistory[ScriptName(symbol)]) == 0 {
		return nil
	}
	startIndex := utils.MomentBinarySearch(shareHistory[ScriptName(symbol)], from)
	endIndex := utils.MomentBinarySearch(shareHistory[ScriptName(symbol)], to)

	requestedRange := shareHistory[ScriptName(symbol)][startIndex:endIndex]
	if len(requestedRange) == 0 {
		return nil
	}
	startPrice := requestedRange[0].Close
	for i, _ := range requestedRange {
		requestedRange[i].PercentChange = ((requestedRange[i].Close - startPrice) / startPrice) * 100
	}
	return requestedRange
}

func GetPriceMFPositionsInTimeRange(symbol string, from, to time.Time) []models.MFHoldingsData {
	requestedRange := mutualFundsTradebook.MutualFundsTradebook[ISIN(symbol)]
	if len(requestedRange) == 0 {
		return nil
	}
	var holdings []models.MFHoldingsData
	totalValue := float64(0)
	totalUnitsHeld := float64(0)
	for _, trade := range requestedRange {
		var transaction float64
		if trade.TradeType == "sell" {
			totalValue -= trade.Price * trade.Quantity
			totalUnitsHeld -= trade.Quantity
			transaction = trade.Price * (-trade.Quantity)
		} else {
			totalValue += trade.Price * trade.Quantity
			totalUnitsHeld += trade.Quantity
			transaction = trade.Price * trade.Quantity
		}
		holdings = append(holdings, models.MFHoldingsData{
			Timestamps:     trade.TradeDate,
			TotalValue:     totalValue,
			TotalUnitsHeld: math.Ceil(totalUnitsHeld),
			Transaction:    transaction,
		})
	}
	return holdings
}

func getCAGR(isin ISIN, from, to time.Time) float64 {
	priceHistory := mutualFundsHistory[isin]
	if len(priceHistory) == 0 {
		return 0
	}
	var cagr float64
	startIndex := utils.MomentBinarySearch(priceHistory, from)
	startPrice := priceHistory[startIndex].GetPrice()
	startTime := priceHistory[startIndex].GetTime()
	endPrice := priceHistory[len(priceHistory)-1].GetPrice()
	endTime := to
	if startIndex != len(priceHistory)-1 {
		periods := time.Duration(endTime.Sub(startTime)).Hours() / 8766
		cagr = (math.Pow((endPrice/startPrice), (1/float64(periods))) - 1) * 100
	}
	return cagr
}

func getXIRR(isin ISIN, from, to time.Time, currentValue float64) float64 {
	tradeHistory := mutualFundsTradebook.MutualFundsTradebook[isin]
	startIndex := utils.MomentBinarySearch(tradeHistory, from)
	array := []MutualFundsTrade{}
	for _, v := range tradeHistory[startIndex:] {
		if v.TradeType == "sell" {
			array = append(array, MutualFundsTrade{
				TradeDate: v.GetTime(),
				Price:     -v.GetPrice(),
			})
			continue
		}
		array = append(array, v)
	}
	array = append(array, MutualFundsTrade{
		TradeDate: to,
		Price:     -currentValue,
	})
	xirr := utils.GetXIRR(array)
	return xirr
}

func GetMFSummmary(from, to time.Time) []models.MFSummary {
	var summary []models.MFSummary
	for isin, trades := range mutualFundsTradebook.MutualFundsTradebook {
		fmt.Println("starting calculation for ", string(mutualFundsTradebook.GetFundNameFromISIN(isin)))
		if len(trades) == 0 {
			continue
		}
		var heldUnits float64
		var moneyInvested float64
		var moneyHarvested float64
		var holdingSince *time.Time
		var currentInvested float64
		var lastInvestment time.Time
		for _, t := range trades {
			if t.TradeType == "buy" {
				if heldUnits < 1 {
					k := t.TradeDate
					holdingSince = &k
					currentInvested = 0
				}
				heldUnits += t.Quantity
				moneyInvested += t.Quantity * t.Price
				currentInvested += t.Quantity * t.Price
				lastInvestment = t.TradeDate
			} else {
				heldUnits -= t.Quantity
				moneyHarvested += t.Quantity * t.Price
				currentInvested -= t.Quantity * t.Price
				if heldUnits < 1 {
					holdingSince = nil
					currentInvested = 0
				}
			}
		}

		priceHistory := mutualFundsHistory[isin]
		if len(priceHistory) == 0 {
			log.Printf("unable to compute summary, price history not found for %s", isin)
			continue
		}
		currentPrice := float64(priceHistory[len(priceHistory)-1].Price)

		var holdingSinceDuration time.Duration
		if holdingSince != nil {
			k := time.Since(*holdingSince)
			h := time.Duration(k.Seconds())
			holdingSinceDuration = h
		} else {
			holdingSinceDuration = 0
		}
		currentValue := math.Ceil(heldUnits) * currentPrice
		var cagr, xirr float64
		if holdingSince != nil {
			cagr = getCAGR(isin, *holdingSince, time.Now())
			xirr = getXIRR(isin, *holdingSince, time.Now(), currentValue)
		}
		fmt.Println(string(mutualFundsTradebook.GetFundNameFromISIN(isin)), cagr, xirr)
		s := models.MFSummary{
			Name:                  string(mutualFundsTradebook.GetFundNameFromISIN(isin)),
			ISIN:                  string(isin),
			HoldingSince:          holdingSinceDuration,
			HoldingFrom:           time.Duration(lastInvestment.Sub(trades[0].TradeDate).Seconds()),
			CurrentValue:          currentValue,
			InvestedValue:         currentInvested,
			AllTimeAbsoluteReturn: currentValue - currentInvested,
			LastInvestment:        time.Duration(time.Since(lastInvestment).Seconds()),
			CAGR:                  cagr,
			XIRR:                  xirr,
		}
		if (currentValue-currentInvested) != 0 && currentInvested != 0 {
			s.AllTimeAbsoluteReturnPercentage = ((currentValue - currentInvested) / currentInvested) * 100
		}
		summary = append(summary, s)
	}
	return summary
}

func GetPriceMFTrendInTimeRange(symbol string, from, to time.Time) []models.MFPriceData {
	if len(mutualFundsHistory[ISIN(symbol)]) == 0 {
		return nil
	}
	startIndex := utils.MomentBinarySearch(mutualFundsHistory[ISIN(symbol)], from)
	endIndex := utils.MomentBinarySearch(mutualFundsHistory[ISIN(symbol)], to)

	requestedRange := mutualFundsHistory[ISIN(symbol)][startIndex:endIndex]
	if len(requestedRange) == 0 {
		return nil
	}
	startPrice := requestedRange[0].Price
	for i, _ := range requestedRange {
		requestedRange[i].PercentChange = ((requestedRange[i].Price - startPrice) / startPrice) * 100
	}
	return requestedRange
}

func GetGrowthComparison(symbols []string, from, to time.Time) []map[string]interface{} {
	growthMap := make(map[time.Time]map[string]float32)
	for _, symbol := range symbols {
		trend := GetPriceTrendInTimeRange(symbol, from, to)
		for _, v := range trend {
			if _, ok := growthMap[v.Timestamps]; !ok {
				growthMap[v.Timestamps] = make(map[string]float32)
				for _, s := range symbols {
					//  init empty value with 0 because some stocks might have started later in the requested time period
					growthMap[v.Timestamps][s] = 0
				}
			}
			growthMap[v.Timestamps][symbol] = v.PercentChange
		}
	}

	response := make([]map[string]interface{}, len(growthMap))
	index := 0
	for timeStamp, mapSymbolToPrice := range growthMap {
		response[index] = make(map[string]interface{})
		for s, p := range mapSymbolToPrice {
			response[index][s] = p
		}
		response[index]["time"] = timeStamp
		index++
	}
	return response
}

func GetMFGrowthComparison(symbols []string, from, to time.Time) []map[string]interface{} {
	growthMap := make(map[time.Time]map[string]float32)
	for _, symbol := range symbols {
		trend := GetPriceMFTrendInTimeRange(symbol, from, to)
		for _, v := range trend {
			if _, ok := growthMap[v.Timestamps]; !ok {
				growthMap[v.Timestamps] = make(map[string]float32)
				for _, s := range symbols {
					//  init empty value with 0 because some stocks might have started later in the requested time period
					growthMap[v.Timestamps][s] = 0
				}
			}
			growthMap[v.Timestamps][symbol] = v.PercentChange
		}
	}

	response := make([]map[string]interface{}, len(growthMap))
	index := 0
	for timeStamp, mapSymbolToPrice := range growthMap {
		response[index] = make(map[string]interface{})
		for s, p := range mapSymbolToPrice {
			fundName := mutualFundsTradebook.GetFundNameFromISIN(ISIN(s)).String()
			response[index][fundName] = p
		}
		response[index]["time"] = timeStamp
		index++
	}
	return response
}
