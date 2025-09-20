package tickerTape_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Mryashbhardwaj/marketAnalysis/core/clients/tickerTape"
)

func TestGetMFSummary(t *testing.T) {

	t.Run("get MF summary data from tickertape", func(t *testing.T) {
		symbol := tickerTape.TtSymbol("M_HDCEQ")
		summary, err := symbol.GetMFSummary()
		assert.Nil(t, err)
		assert.Equal(t, "INF179K01UT0", summary.Isin)
	})
}
