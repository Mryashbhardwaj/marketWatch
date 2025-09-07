package routes

import (
	"github.com/Mryashbhardwaj/marketAnalysis/internal/api/handlers"
	"github.com/gorilla/mux"
)

func SetupRouter(handler *handlers.Handler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/equity/list", handler.GetEquityList).Methods("GET")
	router.HandleFunc("/api/equity/trend", handler.GetTrend).Methods("GET")
	router.HandleFunc("/api/equity/trend/compare", handler.GetTrendComparison).Methods("GET")
	router.HandleFunc("/api/equity/history/refresh", handler.RefreshPriceHistory).Methods("GET")
	router.HandleFunc("/api/equity/breakdown", handler.GetEqBreakdown).Methods("GET")

	router.HandleFunc("/api/mutual_funds/list", handler.GetMutualFundsList).Methods("GET")
	router.HandleFunc("/api/mutual_funds/positions", handler.GetMFPositions).Methods("GET")
	router.HandleFunc("/api/mutual_funds/trend", handler.GetMFTrend).Methods("GET")
	router.HandleFunc("/api/mutual_funds/summary", handler.GetMFSummary).Methods("GET")
	router.HandleFunc("/api/mutual_funds/trend/compare", handler.GetMFGrowthComparison).Methods("GET")
	router.HandleFunc("/api/mutual_funds/history/refresh", handler.RefreshMFPriceHistory).Methods("GET")

	return router
}
