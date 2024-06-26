package tradebook_service

import (
	"fmt"
	"strconv"
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
		tradeTime, err := time.Parse(time.DateOnly, trade.TradeDate)
		if err != nil {
			fmt.Println(err.Error())
		}

		price, err := strconv.ParseFloat(trade.Price, 64)
		if err != nil {
			fmt.Println(err.Error())
		}

		quantity, err := strconv.ParseFloat(trade.Quantity, 64)
		if err != nil {
			fmt.Println(err.Error())
		}
		if trade.TradeType == "sell" {
			totalValue -= price * quantity
			totalUnitsHeld -= quantity
			transaction = price * (-quantity)
		} else {
			totalValue += price * quantity
			totalUnitsHeld += quantity
			transaction = price * quantity
		}
		holdings = append(holdings, models.MFHoldingsData{
			Timestamps:     tradeTime,
			TotalValue:     totalValue,
			TotalUnitsHeld: totalUnitsHeld,
			Transaction:    transaction,
		})
	}
	return holdings
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
