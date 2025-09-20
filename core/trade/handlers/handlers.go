package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Mryashbhardwaj/marketAnalysis/core/trade/models"
	"github.com/Mryashbhardwaj/marketAnalysis/core/trade/service"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/utils"
)

// all methos responses should be properly defined
type Tradebook interface {
	GetMFSummmary(from, to time.Time) []models.MFSummary
	GetPriceMFPositionsInTimeRange(symbol string, from time.Time, to time.Time) []models.MFHoldingsData
	GetMutualFundsList() map[service.FundName]service.ISIN
	GetEquityList() []service.ScriptName
	GetEqBreakdown(symbol string) (service.BreakdownResponse, error)
}

type EquityTrendCache interface {
	GetPriceTrendInTimeRange(symbol string, from time.Time, to time.Time) []models.EquityPriceData
	GetGrowthComparison(symbols []string, from, to time.Time) []map[string]interface{}
	BuildPriceHistoryCache(allShares []service.ScriptName) error
}

type MFTrendCache interface {
	GetPriceMFTrendInTimeRange(symbol string, from time.Time, to time.Time) []models.MFPriceData
	GetMFGrowthComparison(symbols []string, from, to time.Time) []map[string]interface{}
	BuildMFPriceHistoryCache(map[service.FundName]service.ISIN) error
}

type Handler struct {
	logger           *slog.Logger
	tradebookService Tradebook
	equityTrendCache EquityTrendCache
	mfTrendCache     MFTrendCache
}

func GetHandler(tradebookService Tradebook, equityTrendCache EquityTrendCache, mfTrendCache MFTrendCache) *Handler {
	return &Handler{
		logger:           slog.Default(),
		tradebookService: tradebookService,
		equityTrendCache: equityTrendCache,
		mfTrendCache:     mfTrendCache,
	}
}

func (h Handler) GetTrend(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, h.equityTrendCache.GetPriceTrendInTimeRange(symbol, from, to))
}

func (h Handler) GetMFSummary(w http.ResponseWriter, r *http.Request) {
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, h.tradebookService.GetMFSummmary(from, to))
}

func (h Handler) GetMFTrend(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, h.mfTrendCache.GetPriceMFTrendInTimeRange(symbol, from, to))
}

func (h Handler) GetMFPositions(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, h.tradebookService.GetPriceMFPositionsInTimeRange(symbol, from, to))
}

func (h Handler) GetTrendComparison(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")

	symbols := strings.Split(symbol[1:len(symbol)-1], ",")

	if len(symbols) < 2 {
		utils.RespondWithJSON(w, 200, map[string]interface{}{})
		return
	}

	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondWithJSON(w, 200, h.equityTrendCache.GetGrowthComparison(symbols, from, to))
}

// cleaned up
func (h Handler) GetMFGrowthComparison(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("symbol")
	if raw == "" {
		http.Error(w, "missing symbol param", http.StatusBadRequest)
		return
	}
	cleaned := strings.Trim(raw, "{}")

	symbols := strings.Split(cleaned, ",")

	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, h.mfTrendCache.GetMFGrowthComparison(symbols, from, to))
}

func (h Handler) GetMutualFundsList(w http.ResponseWriter, r *http.Request) {
	mfMap := h.tradebookService.GetMutualFundsList()
	var fundList []string
	for fundName, insin := range mfMap {
		fundList = append(fundList, fmt.Sprintf("%s:%s", fundName, insin))
	}
	utils.RespondWithJSON(w, 200, fundList)
}

func (h Handler) GetEquityList(w http.ResponseWriter, r *http.Request) {
	eqList := h.tradebookService.GetEquityList()
	utils.RespondWithJSON(w, 200, eqList)
}

// make server object and get logger and config in handler
func (h Handler) RefreshPriceHistory(w http.ResponseWriter, r *http.Request) {
	allShares := h.tradebookService.GetEquityList()
	err := h.equityTrendCache.BuildPriceHistoryCache(allShares)
	if err != nil {
		utils.RespondWithJSON(w, 505, err.Error())
		return
	}
	utils.RespondWithJSON(w, 200, "Price History Refreshed Successfully")
}

func (h Handler) RefreshMFPriceHistory(w http.ResponseWriter, r *http.Request) {
	allFunds := h.tradebookService.GetMutualFundsList()
	err := h.mfTrendCache.BuildMFPriceHistoryCache(allFunds)
	if err != nil {
		utils.RespondWithJSON(w, 505, err.Error())
		return
	}
	utils.RespondWithJSON(w, 200, "Price History Refreshed Successfully")
}

func (h Handler) GetEqBreakdown(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		utils.RespondWithJSON(w, 400, "Missing 'symbol' parameter")
		return
	}
	breakdown, err := h.tradebookService.GetEqBreakdown(symbol)
	if err != nil {
		utils.RespondWithJSON(w, 500, err.Error())
		return
	}
	utils.RespondWithJSON(w, 200, breakdown)
}
