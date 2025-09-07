package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	MC "github.com/Mryashbhardwaj/marketAnalysis/internal/clients/moneyControl"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/domain/models"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/utils"
	"github.com/pkg/errors"
)

type MutualFundsTrade struct {
	Isin               string
	TradeDate          time.Time
	Exchange           string
	Segment            string
	Series             string
	TradeType          string
	Auction            string
	Quantity           float64
	Price              float64
	TradeID            string
	OrderID            string
	OrderExecutionTime string
}

func (m MutualFundsTrade) GetTime() time.Time {
	return m.TradeDate
}
func (m MutualFundsTrade) GetPrice() float64 {
	return m.Price
}

type FundName string

func (f FundName) String() string {
	return string(f)
}

type ISIN string

type MFTrendCache struct {
	History        map[ISIN][]models.MFPriceData
	ISINToFundName map[ISIN]FundName
	logger         *slog.Logger
}

func GetMFTrendCache(logger *slog.Logger, allFunds map[ISIN]FundName) *MFTrendCache {
	m := &MFTrendCache{
		History:        make(map[ISIN][]models.MFPriceData),
		ISINToFundName: make(map[ISIN]FundName),
		logger:         logger,
	}
	m.ISINToFundName = allFunds

	for isin := range allFunds {
		history, err := buildFundsCacheFromFile(isin)
		if err != nil {
			fmt.Printf("error reading MF cache for %s: %s\n", isin, err)
			continue
		}
		m.History[isin] = history
	}

	return m
}

func (m *MFTrendCache) GetPriceMFTrendInTimeRange(symbol string, from, to time.Time) []models.MFPriceData {
	if len(m.History[ISIN(symbol)]) == 0 {
		return nil
	}
	startIndex := utils.MomentBinarySearch(m.History[ISIN(symbol)], from)
	endIndex := utils.MomentBinarySearch(m.History[ISIN(symbol)], to)

	requestedRange := m.History[ISIN(symbol)][startIndex:endIndex]
	if len(requestedRange) == 0 {
		return nil
	}
	startPrice := requestedRange[0].Price
	for i := range requestedRange {
		requestedRange[i].PercentChange = ((requestedRange[i].Price - startPrice) / startPrice) * 100
	}
	return requestedRange
}

func (m *MFTrendCache) GetMFGrowthComparison(symbols []string, from, to time.Time) []map[string]interface{} {
	growthMap := make(map[time.Time]map[string]float32)
	for _, symbol := range symbols {
		trend := m.GetPriceMFTrendInTimeRange(symbol, from, to)
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
			fundName := m.ISINToFundName[ISIN(s)].String()
			response[index][fundName] = p
		}
		response[index]["time"] = timeStamp
		index++
	}
	return response
}

// func BuildMFTrendCacheIfMissing() error {
// 	_ = os.MkdirAll("./data/trends/MF", os.ModePerm)

// 	for isin, history := range mutualFundsHistory {
// 		if len(history) == 0 {
// 			continue
// 		}

// 		filePath := fmt.Sprintf("./data/trends/MF/%s.json", isin)
// 		if _, err := os.Stat(filePath); err == nil {
// 			continue // file exists, skip
// 		}

// 		startPrice := history[0].Price
// 		for i := range history {
// 			history[i].PercentChange = ((history[i].Price - startPrice) / startPrice) * 100
// 		}

// 		data, err := json.Marshal(history)
// 		if err != nil {
// 			return fmt.Errorf("marshal failed for %s: %w", isin, err)
// 		}

// 		if err := os.WriteFile(filePath, data, os.ModePerm); err != nil {
// 			return fmt.Errorf("write failed for %s: %w", isin, err)
// 		}
// 	}

// 	return nil
// }

func persistMFInFile(symbol string, trend interface{}) error {
	fileContent, err := json.Marshal(trend)
	if err != nil {
		return errors.Wrap(err, "unable to persist MF trade file")
	}
	if _, err := os.Stat("./data/trends/MF/"); os.IsNotExist(err) {
		if err := os.MkdirAll("./data/trends/MF/", os.ModePerm); err != nil {
			return errors.Wrap(err, "unable to create MF trends directory")
		}
	}
	fileName := fmt.Sprintf("./data/trends/MF/%s.json", symbol)
	return os.WriteFile(fileName, fileContent, os.ModePerm)
}

// persist call comes from here for mf
func (m *MFTrendCache) BuildMFPriceHistoryCache(allFunds map[FundName]ISIN) error {
	var errorList []string
	for name, isin := range allFunds {
		history, err := MC.GetMFHistoryFromMoneyControll(string(isin))
		if err != nil {
			fmt.Printf("error fetching history for MF %s, err:%s", name, err.Error())
			errorList = append(errorList, fmt.Sprintf("error fetching history for %s, err:%s", isin, err.Error()))
			continue
		}
		m.History[isin] = history
		err = persistMFInFile(string(isin), history)
		if err != nil {
			fmt.Printf("error persisting history for %s, err:%s", isin, err.Error())
			errorList = append(errorList, fmt.Sprintf("error persisting history for %s, err:%s", isin, err.Error()))
			continue
		}
	}
	if len(errorList) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorList, "\n"))
}

func buildFundsCacheFromFile(symbol ISIN) ([]models.MFPriceData, error) {
	fileName := fmt.Sprintf("./data/trends/MF/%s.json", symbol)
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	trend := []models.MFPriceData{}
	err = json.Unmarshal(fileContent, &trend)
	return trend, err
}
