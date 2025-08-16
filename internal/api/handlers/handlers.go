package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	tradebook_service "github.com/Mryashbhardwaj/marketAnalysis/internal/domain/service"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/utils"
)

type mutualFundOverview struct { //nolint:unused
	Symbol        string
	Name          string
	CurrentTotal  float64
	HoldingsStart time.Time
	HoldingsEnd   time.Time
} // TODO: remove this struct, not used

func GetTrend(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, tradebook_service.GetPriceTrendInTimeRange(symbol, from, to))
}

func GetMFSummary(w http.ResponseWriter, r *http.Request) {
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, tradebook_service.GetMFSummmary(from, to))
}

func GetMFTrend(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, tradebook_service.GetPriceMFTrendInTimeRange(symbol, from, to))
}

func GetMFPositions(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	from, to, err := utils.GetTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.RespondWithJSON(w, 200, tradebook_service.GetPriceMFPositionsInTimeRange(symbol, from, to))
}

func GetTrendComparison(w http.ResponseWriter, r *http.Request) {
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

	utils.RespondWithJSON(w, 200, tradebook_service.GetGrowthComparison(symbols, from, to))
}

// cleaned up
func GetMFGrowthComparison(w http.ResponseWriter, r *http.Request) {
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

	utils.RespondWithJSON(w, http.StatusOK, tradebook_service.GetMFGrowthComparison(symbols, from, to))
}

func GetMutualFundsList(w http.ResponseWriter, r *http.Request) {
	mfList := tradebook_service.GetMutualFundsList()
	utils.RespondWithJSON(w, 200, mfList)
}

func GetEquityList(w http.ResponseWriter, r *http.Request) {
	eqList := tradebook_service.GetEquityList()
	utils.RespondWithJSON(w, 200, eqList)
}

func RefreshPriceHistory(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	err := tradebook_service.BuildPriceHistoryCache()
	elapsed := time.Since(start)
	fmt.Printf("equity cache built in %s", elapsed)
	if err != nil {
		utils.RespondWithJSON(w, 505, err.Error())
		return
	}
	utils.RespondWithJSON(w, 200, "Price History Refreshed Successfully")
}

func RefreshMFPriceHistory(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	err := tradebook_service.BuildMFPriceHistoryCache()
	elapsed := time.Since(start)
	fmt.Printf("mutual funds cache built in %s", elapsed)
	if err != nil {
		utils.RespondWithJSON(w, 505, err.Error())
		return
	}
	utils.RespondWithJSON(w, 200, "Price History Refreshed Successfully")
}

func GetEqBreakdown(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		utils.RespondWithJSON(w, 400, "Missing 'symbol' parameter")
		return
	}
	breakdown, err := tradebook_service.GetEqBreakdown(symbol)
	if err != nil {
		utils.RespondWithJSON(w, 500, err.Error())
		return
	}
	utils.RespondWithJSON(w, 200, breakdown)
}
