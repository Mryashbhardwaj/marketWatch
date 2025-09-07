package service

import (
	"fmt"
	"log"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Mryashbhardwaj/marketAnalysis/internal/domain/models"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/utils"
	"github.com/pkg/errors"
)

type TradeRecord struct {
	Date     string  `json:"date"`
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Type     string  `json:"type"`
}

type BreakdownResponse struct {
	Symbol          string        `json:"symbol"`
	TotalBuyQty     float64       `json:"total_buy_qty"`
	TotalBuyValue   float64       `json:"total_buy_value"`
	TotalSellQty    float64       `json:"total_sell_qty"`
	TotalSellValue  float64       `json:"total_sell_value"`
	NetQuantity     float64       `json:"net_quantity"`
	TotalInvestment float64       `json:"total_investment"`
	TradeHistory    []TradeRecord `json:"trade_history"`
}

type EquityTradebook struct {
	AllShares       []ScriptName
	EquityTradebook map[ScriptName][]EquityTrade
}

type MutualFundsTradebook struct {
	AllFunds             map[FundName]ISIN
	ISINToFundName       map[ISIN]FundName
	MutualFundsTradebook map[ISIN][]MutualFundsTrade
}

type TradebookService struct {
	logger                    *slog.Logger
	EquityTradebookCache      *EquityTradebook
	MutualFundsTradebookCache *MutualFundsTradebook
}

func GetTradebookService(eqTradebookDir, mfTradebookDir string, logger *slog.Logger) (*TradebookService, error) {
	if mfTradebookDir == "" && eqTradebookDir == "" {
		return nil, errors.Errorf("no tradefiles directory set for equity or mutual funds")
	}

	t := &TradebookService{
		logger: logger,
	}
	if mfTradebookDir != "" {
		err := t.BuildMFTradeBook(mfTradebookDir)
		if err != nil {
			logger.Error("failed to get mutual funds tradebook", slog.String("error", err.Error()))
			return nil, err
		}
	}

	if eqTradebookDir != "" {
		err := t.BuildEquityTradeBook(eqTradebookDir)
		if err != nil {
			logger.Error("failed to build equity tradebook", slog.String("error", err.Error()))
			return nil, err
		}
	}

	return t, nil
}

func (t *TradebookService) BuildMFTradeBook(tradebookDir string) error {
	var mutualFundsTradebookCache MutualFundsTradebook
	tradeMap, allFunds, err := readMFTradeFiles(tradebookDir)
	if err != nil {
		return errors.Wrap(err, "unable to read MF trade file")
	}
	mutualFundsTradebookCache.MutualFundsTradebook = tradeMap
	mutualFundsTradebookCache.AllFunds = allFunds
	mutualFundsTradebookCache.ISINToFundName = make(map[ISIN]FundName)
	for k, v := range allFunds {
		mutualFundsTradebookCache.ISINToFundName[v] = k
	}
	t.MutualFundsTradebookCache = &mutualFundsTradebookCache

	return nil
}

// explain the purpose of this function
func readMFTradeFiles(tradebookDir string) (map[ISIN][]MutualFundsTrade, map[FundName]ISIN, error) {
	// to remove duplidate trade ids
	tradeSet := make(map[string]struct{})
	allFunds := make(map[FundName]ISIN)

	tradeFiles, err := utils.ReadDir(tradebookDir)
	if err != nil {
		return nil, nil, err
	}
	tradebookCSV, err := utils.ReadCSV(tradeFiles)
	if err != nil {
		return nil, nil, err
	}
	tradebook := make(map[ISIN][]MutualFundsTrade)
	// avoid magic numbers, instead of 10 it should be index of trade_id
	for _, record := range tradebookCSV {
		if record[0] == "symbol" { //suh handling shoul dnot be needed, instead handle using skip header
			continue
		}
		if _, ok := tradeSet[record[10]]; ok {
			continue
		}
		tradeSet[record[10]] = struct{}{}

		symbol := FundName(record[0])
		isin := ISIN(record[1])
		allFunds[symbol] = isin
		if _, ok := tradebook[isin]; !ok {
			tradebook[isin] = []MutualFundsTrade{}
		}

		tradeTime, err := time.Parse(time.DateOnly, record[2])
		if err != nil {
			return nil, nil, err
		}

		quantityString := record[8]
		quantity, err := strconv.ParseFloat(quantityString, 64)
		if err != nil {
			return nil, nil, err
		}

		priceString := record[9]
		price, err := strconv.ParseFloat(priceString, 64)
		if err != nil {
			return nil, nil, err
		}

		tradebook[isin] = append(tradebook[isin], MutualFundsTrade{
			Isin:               record[1],
			TradeDate:          tradeTime,
			Exchange:           record[3],
			Segment:            record[4],
			Series:             record[5],
			TradeType:          record[6],
			Auction:            record[7],
			Quantity:           quantity,
			Price:              price,
			TradeID:            record[10],
			OrderID:            record[11],
			OrderExecutionTime: record[12],
		})

	}

	for isin := range tradebook {
		sort.Slice(tradebook[isin], func(i, j int) bool {
			return tradebook[isin][i].TradeDate.Before(tradebook[isin][j].TradeDate)
		})
	}

	return tradebook, allFunds, nil
}

func (t *TradebookService) BuildEquityTradeBook(tradebookDir string) error {
	tradeMap, err := readEquityTradeFiles(tradebookDir)
	if err != nil {
		return err
	}

	var trickers []ScriptName
	for fundName := range tradeMap {
		trickers = append(trickers, fundName)
	}

	var EquityTradebookCache EquityTradebook
	EquityTradebookCache.AllShares = trickers
	EquityTradebookCache.EquityTradebook = tradeMap

	t.EquityTradebookCache = &EquityTradebookCache
	return nil
}

func (t *TradebookService) GetMutualFundsList() map[FundName]ISIN {
	return t.MutualFundsTradebookCache.AllFunds
}

func (t *TradebookService) GetEquityList() []ScriptName {
	return t.EquityTradebookCache.AllShares
}

func readEquityTradeFiles(tradebookDir string) (map[ScriptName][]EquityTrade, error) {
	// to remove duplidate trade ids
	tradeSet := make(map[string]struct{})

	tradeFiles, err := utils.ReadDir(tradebookDir)
	if err != nil {
		return nil, err
	}
	tradebookCSV, err := utils.ReadCSV(tradeFiles)
	if err != nil {
		return nil, err
	}
	tradebook := make(map[ScriptName][]EquityTrade)
	for _, record := range tradebookCSV {
		if record[1] == "symbol" {
			continue
		}
		if _, ok := tradeSet[record[10]]; ok {
			continue
		}
		tradeSet[record[10]] = struct{}{}

		symbol := ScriptName(record[0])
		if _, ok := tradebook[symbol]; !ok {
			tradebook[symbol] = []EquityTrade{}
		}
		tradebook[symbol] = append(tradebook[symbol], EquityTrade{
			Symbol:             record[1],
			TradeDate:          record[2],
			Exchange:           record[3],
			Segment:            record[4],
			TradeType:          record[6],
			Quantity:           record[8],
			Price:              record[9],
			OrderExecutionTime: record[12],
		})
	}
	return tradebook, nil
}

func (t *TradebookService) GetPriceMFPositionsInTimeRange(symbol string, from, to time.Time) []models.MFHoldingsData {
	requestedRange := t.MutualFundsTradebookCache.MutualFundsTradebook[ISIN(symbol)]
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

// todo: this should be removed
func (t *TradebookService) getCAGR(isin ISIN, from, to time.Time) float64 {
	priceHistory := t.MutualFundsTradebookCache.MutualFundsTradebook[isin]
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

// todo: this should be removed
func (t *TradebookService) getXIRR(isin ISIN, from, to time.Time, currentValue float64) float64 {
	tradeHistory := t.MutualFundsTradebookCache.MutualFundsTradebook[isin]
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

func (t *TradebookService) GetMFSummmary(from, to time.Time) []models.MFSummary {
	var summary []models.MFSummary
	for isin, trades := range t.MutualFundsTradebookCache.MutualFundsTradebook {
		fmt.Println("starting calculation for ", string(t.GetFundNameFromISIN(isin)))
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

		priceHistory := t.MutualFundsTradebookCache.MutualFundsTradebook[isin]
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
			cagr = t.getCAGR(isin, *holdingSince, time.Now())
			xirr = t.getXIRR(isin, *holdingSince, time.Now(), currentValue)
		}
		fmt.Println(string(t.GetFundNameFromISIN(isin)), cagr, xirr)
		s := models.MFSummary{
			Name:                  string(t.GetFundNameFromISIN(isin)),
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

func (t *TradebookService) GetFundNameFromISIN(k ISIN) FundName {
	return t.MutualFundsTradebookCache.ISINToFundName[k]
}

func (t *TradebookService) GetEqBreakdown(symbol string) (BreakdownResponse, error) {
	script := ScriptName(strings.ToUpper(symbol))
	trades, ok := t.EquityTradebookCache.EquityTradebook[script]
	if !ok {
		return BreakdownResponse{}, fmt.Errorf("no data for symbol: %s", symbol)
	}

	var (
		buyQty, sellQty, buyValue, sellValue float64
		history                              []TradeRecord
	)

	for _, trade := range trades {
		price, err1 := strconv.ParseFloat(trade.Price, 64)
		qty, err2 := strconv.ParseFloat(trade.Quantity, 64)
		if err1 != nil || err2 != nil {
			continue
		}

		record := TradeRecord{
			Date:     trade.TradeDate,
			Price:    price,
			Quantity: qty,
			Type:     strings.ToLower(trade.TradeType),
		}
		history = append(history, record)

		switch record.Type {
		case "buy":
			buyQty += qty
			buyValue += qty * price
		case "sell":
			sellQty += qty
			sellValue += qty * price
		}
	}

	netQty := buyQty - sellQty
	avgBuy := 0.0
	totalInvestment := 0.0
	if netQty > 0 {
		avgBuy = buyValue / buyQty
		totalInvestment = netQty * avgBuy
	}

	return BreakdownResponse{
		Symbol:          symbol,
		TotalBuyQty:     buyQty,
		TotalBuyValue:   buyValue,
		TotalSellQty:    sellQty,
		TotalSellValue:  sellValue,
		NetQuantity:     netQty,
		TotalInvestment: totalInvestment,
		TradeHistory:    history,
	}, nil
}
