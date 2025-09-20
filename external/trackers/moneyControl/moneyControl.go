package moneycontrol

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/Mryashbhardwaj/marketAnalysis/core/trade/models"
)

func GetMFHistoryFromMoneyControll(isin string) ([]models.MFPriceData, error) {
	priceAPIURL := fmt.Sprintf("https://www.moneycontrol.com/mc/widget/mfnavonetimeinvestment/get_chart_value?isin=%s&dur=ALL", isin)
	fmt.Println(priceAPIURL)

	req, err := http.NewRequest("GET", priceAPIURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("error closing response body: %v\n", cerr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	k := models.MoneyControlMFHistoryResponse{}
	err = json.Unmarshal(body, &k)

	priceHistory := make([]models.MFPriceData, len(k.Trend))

	for i, v := range k.Trend {
		date, _ := time.Parse(time.DateOnly, v.Date)
		priceHistory[i].Timestamps = date
		priceHistory[i].Price = v.Price
	}
	return priceHistory, err
}

func GetEQHistoryFromMoneyControll(tickerSymbol string) (*models.MoneyControlResponse, error) {
	startTime := time.Unix(490147200, 0)
	endTime := time.Now()
	durationSince := math.Ceil(endTime.Sub(startTime).Hours() / 24)
	priceAPIURL := fmt.Sprintf("https://priceapi.moneycontrol.com/techCharts/indianMarket/stock/history?symbol=%s&resolution=1D&from=%d&to=%d&countback=%.f&currencyCode=INR", tickerSymbol, startTime.Unix(), endTime.Unix(), durationSince)
	fmt.Println(priceAPIURL)

	req, err := http.NewRequest("GET", priceAPIURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("error closing response body: %v\n", cerr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	k := models.MoneyControlResponse{}
	err = json.Unmarshal(body, &k)
	return &k, err
}
