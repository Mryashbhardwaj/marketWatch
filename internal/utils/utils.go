package utils

import (
	"encoding/csv"
	"encoding/json"
	"path"

	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func ReadDir(dir string) ([]string, error) {
	tradeFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "doesNotExist")
	}
	var tradeFilesStrings []string
	for _, tf := range tradeFiles {
		tradeFilesStrings = append(tradeFilesStrings, path.Join(dir, tf.Name()))
	}
	return tradeFilesStrings, nil
}

// skip header - flag - int - skip how many lines
// change name to readandmergecsvfiles
func ReadCSV(tradeFiles []string) ([][]string, error) {
	var tradeFilesCombined [][]string
	for _, tf := range tradeFiles {
		file, err := os.ReadFile(tf)
		if err != nil {
			return nil, err
		}
		r := csv.NewReader(strings.NewReader(string(file)))
		k, err := r.ReadAll()
		if err != nil {
			return nil, err
		}
		tradeFilesCombined = append(tradeFilesCombined, k...)
	}
	return tradeFilesCombined, nil
}

func GetTimeRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" {
		fromStr = "490147200000"
	}
	if toStr == "" {
		toStr = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	// Parse the "from" timestamp
	fromMilli, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid 'from' timestamp format. Expected epoch milliseconds")
	}
	from := time.Unix(fromMilli/1000, 0)

	// Parse the "to" timestamp
	toMilli, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid 'to' timestamp format. Expected epoch milliseconds")
	}
	to := time.Unix(toMilli/1000, 0)

	return from, to, nil
}

// todo: accept logger as an argument and log using structured logging
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(response); err != nil {
		log.Println("error writing response:", err)
	}
}

type TimeGetter interface {
	GetTime() time.Time
}

type TradesGetter interface {
	GetTime() time.Time
	GetPrice() float64
}

func GetCAGR[V TradesGetter](trades []V) float64 {
	if len(trades) == 0 {
		return 0
	}
	beginningValue := trades[0].GetPrice()
	endingValue := trades[len(trades)-1].GetPrice()

	beginningTime := trades[0].GetTime()
	endingTime := trades[len(trades)-1].GetTime()
	periods := time.Duration(endingTime.Sub(beginningTime)).Hours() / 8766
	if periods <= 0 {
		return 0
	}
	return (math.Pow((endingValue/beginningValue), (1/float64(periods))) - 1) * 100
}

func GetXIRR[V TradesGetter](trades []V) float64 {
	if len(trades) == 0 {
		return 0
	}
	// there are moments when all the units are sold, i want to calculate the XIRR since that time
	// its a design choice
	return calculateXIRR(trades)
}

func calculateXIRR[V TradesGetter](cashFlows []V) float64 {
	guess := 0.5
	const maxIterations = 1000
	const tolerance = 1e-6

	// Convert dates to years from the first date
	daysInYear := 365.25
	baseDate := cashFlows[0].GetTime()
	years := make([]float64, len(cashFlows))
	for i, cf := range cashFlows {
		years[i] = cf.GetTime().Sub(baseDate).Hours() / 24.0 / daysInYear
	}

	// Define the XIRR function and its derivative
	f := func(x float64) float64 {
		result := 0.0
		for i, cf := range cashFlows {
			result += cf.GetPrice() / math.Pow(1.0+x, years[i])
		}
		return result
	}

	fPrime := func(x float64) float64 {
		result := 0.0
		for i, cf := range cashFlows {
			result -= years[i] * cf.GetPrice() / math.Pow(1.0+x, years[i]+1)
		}
		return result
	}

	// Use Newton-Raphson method to find the root
	rate := guess
	for i := 0; i < maxIterations; i++ {
		fValue := f(rate)
		fDerivative := fPrime(rate)
		if math.Abs(fDerivative) < tolerance {
			break
		}
		newRate := rate - fValue/fDerivative
		if math.Abs(newRate-rate) < tolerance {
			return newRate
		}
		rate = newRate
	}
	log.Println("unable to converge")
	return 0
}

func MomentBinarySearch[V TimeGetter](timestamps []V, target time.Time) int {
	left, right := 0, len(timestamps)-1
	nearestIndex := -1
	minDiff := math.MaxInt64

	for left <= right {
		mid := left + (right-left)/2

		// Check if the target is present at mid
		if timestamps[mid].GetTime().Equal(target) {
			return mid
		}

		// Update the nearest index if the current difference is smaller
		diff := absDuration(timestamps[mid].GetTime().Sub(target))
		if diff < time.Duration(minDiff) {
			minDiff = int(diff)
			nearestIndex = mid
		}

		// If the target is greater, ignore the left half
		if timestamps[mid].GetTime().Before(target) {
			left = mid + 1
		} else {
			// If the target is smaller, ignore the right half
			right = mid - 1
		}
	}
	return nearestIndex
}

// absDuration is a helper function to calculate the absolute value of a time.Duration.
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// // todo: only use this fucntion
// func persistTrendInFile(fileName string, trend interface{}) error {
// 	fileContent, err := json.Marshal(trend)
// 	if err != nil {
// 		return err
// 	}
// 	dirName := path.Base(fileName)

// 	if _, err := os.Stat(dirName); os.IsNotExist(err) {
// 		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
// 			return errors.Wrap(err, "unable to create trends directory for "+dirName)
// 		}
// 	}

// 	return os.WriteFile(fileName, fileContent, os.ModePerm)
// }
