package api

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedPrecision(t *testing.T) {
	tests := []struct {
		name      string
		num       float64
		precision int
		expected  float64
	}{
		{
			name:      "precision = 2 with input of 4 points after the decimal",
			num:       303.6800,
			precision: 2,
			expected:  303.68,
		},
		{
			name:      "round up for precision",
			num:       100.56700,
			precision: 2,
			expected:  100.57,
		},
		{
			name:      "precision = 2 with input of 1 point1 after the decimal",
			num:       100.5,
			precision: 2,
			expected:  100.50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := FixedPrecision(tt.num, tt.precision)

			assert.Equal(t, res, tt.expected)
		})
	}
}

func Test_sanitize(t *testing.T) {

	tests := []struct {
		name       string
		bodyReader io.ReadCloser
		nDays      int
		prices     []*DailyPrice
		avgClose   float64
		err        string
	}{
		{
			name: "transforms and orders",
			bodyReader: ioutil.NopCloser(strings.NewReader(`{
    "Meta Data": {
        "1. Information": "Daily Prices (open, high, low, close) and Volumes",
        "2. Symbol": "IBM",
        "3. Last Refreshed": "2022-04-01",
        "4. Output Size": "Compact",
        "5. Time Zone": "US/Eastern"
    },
    "Time Series (Daily)": {
        "2022-04-01": {
            "1. open": "129.6600",
            "2. high": "130.2700",
            "3. low": "128.0600",
            "4. close": "130.1500",
            "5. volume": "4012373"
        },
        "2022-03-31": {
            "1. open": "130.7200",
            "2. high": "131.8800",
            "3. low": "130.0000",
            "4. close": "130.0200",
            "5. volume": "4274029"
        },
        "2022-03-30": {
            "1. open": "132.0100",
            "2. high": "133.0800",
            "3. low": "131.3900",
            "4. close": "132.1300",
            "5. volume": "2622860"
        }
    }
}`)),
			nDays: 3,
			prices: []*DailyPrice{
				{
					Day: "2022-04-01",
					Price: &Price{
						Open:   "129.6600",
						High:   "130.2700",
						Low:    "128.0600",
						Close:  "130.1500",
						Volume: "4012373",
					},
				},
				{
					Day: "2022-03-31",
					Price: &Price{
						Open:   "130.7200",
						High:   "131.8800",
						Low:    "130.0000",
						Close:  "130.0200",
						Volume: "4274029",
					},
				},
				{
					Day: "2022-03-30",
					Price: &Price{
						Open:   "132.0100",
						High:   "133.0800",
						Low:    "131.3900",
						Close:  "132.1300",
						Volume: "2622860",
					},
				},
			},
			avgClose: 130.77,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices, avgClose, err := sanitize(tt.bodyReader, tt.nDays)
			if err != nil {
				assert.Equal(t, err.Error(), tt.err)

				return
			}

			assert.Equalf(t, tt.prices, prices, "sanitize(%v, %v)", tt.bodyReader, tt.nDays)
			assert.Equalf(t, tt.avgClose, avgClose, "sanitize(%v, %v)", tt.bodyReader, tt.nDays)
		})
	}
}
