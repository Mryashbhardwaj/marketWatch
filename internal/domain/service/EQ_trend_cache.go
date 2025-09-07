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

type ScriptName string

func (s ScriptName) String() string {
	return string(s)
}

type EquityTrade struct {
	Isin               string
	Symbol             string
	TradeDate          string
	Exchange           string
	Segment            string
	Series             string
	TradeType          string
	Auction            string
	Quantity           string
	Price              string
	TradeId            string
	OrderId            string
	OrderExecutionTime string
}

type EquityTrendCache struct {
	History map[ScriptName][]models.EquityPriceData
	logger  *slog.Logger
}

func GetEquityTrendCache(logger *slog.Logger, allShares []ScriptName) *EquityTrendCache {
	history := BuildEquityPriceHistoryCacheFromFile(allShares)
	if history == nil {
		history = make(map[ScriptName][]models.EquityPriceData)
	}

	marketTrendCache := &EquityTrendCache{
		History: history,
		logger:  logger,
	}
	return marketTrendCache
}

func (e *EquityTrendCache) GetPriceTrendInTimeRange(symbol string, from, to time.Time) []models.EquityPriceData {
	if len(e.History[ScriptName(symbol)]) == 0 {
		return nil
	}
	startIndex := utils.MomentBinarySearch(e.History[ScriptName(symbol)], from)
	endIndex := utils.MomentBinarySearch(e.History[ScriptName(symbol)], to)

	requestedRange := e.History[ScriptName(symbol)][startIndex:endIndex]
	if len(requestedRange) == 0 {
		return nil
	}
	startPrice := requestedRange[0].Close
	for i := range requestedRange {
		requestedRange[i].PercentChange = ((requestedRange[i].Close - startPrice) / startPrice) * 100
	}
	return requestedRange
}

// todo: make this function an object function
func (e *EquityTrendCache) BuildPriceHistoryCache(allShares []ScriptName) error {
	var errorList []string
	for _, symbol := range allShares {
		history, err := fetchTradeHistories(symbol)
		if err != nil {
			fmt.Printf("error fetching history for %s, err:%s", symbol, err.Error())
			errorList = append(errorList, fmt.Sprintf("error fetching history for %s, err:%s", symbol, err.Error()))
			continue
		}
		e.History[symbol] = history
		err = persistInFile(string(symbol), history)
		if err != nil {
			fmt.Printf("error persisting history for %s, err:%s", symbol, err.Error())
			errorList = append(errorList, fmt.Sprintf("error persisting history for %s, err:%s", symbol, err.Error()))
			continue
		}
	}
	if len(errorList) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorList, "\n"))
}

func (e *EquityTrendCache) GetGrowthComparison(symbols []string, from, to time.Time) []map[string]interface{} {
	// This holds, for each timestamp, how much each symbol has changed
	growthMap := make(map[time.Time]map[string]float32)
	for _, symbol := range symbols {
		trend := e.GetPriceTrendInTimeRange(symbol, from, to)
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

func BuildEquityPriceHistoryCacheFromFile(allShares []ScriptName) map[ScriptName][]models.EquityPriceData {
	shareHistory := make(map[ScriptName][]models.EquityPriceData)
	for _, symbol := range allShares {
		history, err := buildEquityCacheFromFile(symbol)
		if err != nil {
			fmt.Printf("error fetching history from file for %s, err:%s\n", symbol, err.Error())
			continue
		}
		shareHistory[symbol] = history
	}
	return shareHistory
}

func buildEquityCacheFromFile(symbol ScriptName) ([]models.EquityPriceData, error) {
	fileName := fmt.Sprintf("./data/trends/EQ/%s.json", symbol)
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	trend := []models.EquityPriceData{}
	err = json.Unmarshal(fileContent, &trend)
	return trend, err
}

func persistInFile(symbol string, trend interface{}) error {
	fileContent, err := json.Marshal(trend)
	if err != nil {
		return err
	}
	if _, err := os.Stat("./data/trends/EQ/"); os.IsNotExist(err) {
		if err := os.MkdirAll("./data/trends/EQ/", os.ModePerm); err != nil {
			return errors.Wrap(err, "unable to create EQ trends directory")
		}
	}
	fileName := fmt.Sprintf("./data/trends/EQ/%s.json", symbol)
	return os.WriteFile(fileName, fileContent, os.ModePerm)
}

func fetchTradeHistories(script ScriptName) ([]models.EquityPriceData, error) {
	k, err := MC.GetEQHistoryFromMoneyControll(script.String())
	if err != nil {
		return nil, err
	}

	candlePoints := make([]models.EquityPriceData, len(k.T))

	for i, timeStamp := range k.T {
		candlePoints[i] = models.EquityPriceData{
			Close:      k.C[i],
			High:       k.H[i],
			Volume:     k.V[i],
			Open:       k.O[i],
			Low:        k.L[i],
			Timestamps: time.Unix(timeStamp, 0),
		}
	}
	return candlePoints, nil
}
