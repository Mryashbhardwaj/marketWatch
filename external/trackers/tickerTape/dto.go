package tickerTape

type MFSummaryResp struct {
	Data struct {
		Meta MFSummary `json:"meta"`
	} `json:"data"`
}

type MFSummary struct {
	Name               string `json:"name"`
	Isin               string `json:"isin"`
	Plan               string `json:"plan"`
	Option             string `json:"option"`
	Amc                string `json:"amc"`
	Visible            bool   `json:"visible"`
	Active             bool   `json:"active"`
	RiskClassification string `json:"riskClassification"`
	Subsector          string `json:"subsector"`
	SubsectorDesc      string `json:"subsectorDesc"`
	FullName           string `json:"fullName"`
	FundType           string `json:"fundType"`
}
