package models

type MFHistoryMC struct {
	Date  string  `json:"navDate"`
	Price float32 `json:"navValue"`
}

type MoneyControlMFHistoryResponse struct {
	Trend []MFHistoryMC `json:"g1"`
}

type MoneyControlResponse struct {
	C []float32
	H []float32
	L []float32
	O []float32
	T []int64
	V []float32
}
