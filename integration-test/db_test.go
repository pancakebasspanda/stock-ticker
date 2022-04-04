package integration_test

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"stock_ticker/api"
	"stock_ticker/storage"
	"testing"
)

func TestRedis_AddPrices(t *testing.T) {

	// run docker-compose up redis so that localhost version of redis is up
	reJsonHandler := rejson.NewReJSONHandler()

	// Redigo Client
	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialPassword(""))
	if err != nil {
		t.Errorf("test error :%e", err)
	}

	reJsonHandler.SetRedigoClient(conn)

	tests := []struct {
		name   string
		rh     *rejson.Handler
		prices *api.JSONResponse
		err    string
	}{
		{
			name: "successfully add and save stock prices",
			rh:   reJsonHandler,
			prices: &api.JSONResponse{
				MetaData: api.MD{
					Information:   "Daily Prices (open, high, low, close) and Volumes\"",
					Symbol:        "IBM",
					LastRefreshed: "2022-04-01",
					OutputSize:    "Compact",
					TimeZone:      "US/Eastern",
				},
				DailyPrices: map[string]api.Price{
					"Test - 2022-04-01": {
						Open:   "309.3700",
						High:   "310.1300",
						Low:    "305.5400",
						Close:  "309.4200",
						Volume: "27110529",
					},
					"Test - 2022-03-31": {
						Open:   "313.9000",
						High:   "315.1400",
						Low:    "307.8900",
						Close:  "308.3100",
						Volume: "33422070",
					},
					"Test - 2022-03-30": {
						Open:   "313.7600",
						High:   "315.9500",
						Low:    "311.5800",
						Close:  "313.8600",
						Volume: "28163555",
					},
				},
			},
			err: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &storage.Redis{
				Rh: tt.rh,
			}

			if err := r.AddPrices(tt.prices); err != nil {
				assert.Equal(t, err.Error(), tt.err)

				return
			}

			assert.Equal(t, &api.Price{
				Open:   "309.3700",
				High:   "310.1300",
				Low:    "305.5400",
				Close:  "309.4200",
				Volume: "27110529",
			}, getSpecificPrice(t, tt.rh, "Test - 2022-04-01"))

			assert.Equal(t, &api.Price{
				Open:   "313.9000",
				High:   "315.1400",
				Low:    "307.8900",
				Close:  "308.3100",
				Volume: "33422070",
			}, getSpecificPrice(t, tt.rh, "Test - 2022-03-31"))

			assert.Equal(t, &api.Price{
				Open:   "313.7600",
				High:   "315.9500",
				Low:    "311.5800",
				Close:  "313.8600",
				Volume: "28163555",
			}, getSpecificPrice(t, tt.rh, "Test - 2022-03-30"))

		})

	}
}

func TestRedis_GetPriceInfo(t *testing.T) {

	// run docker-compose up redis so that localhost version of redis is up
	reJsonHandler := rejson.NewReJSONHandler()

	// Redigo Client
	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialPassword(""))
	if err != nil {
		t.Errorf("test error :%e", err)
	}

	reJsonHandler.SetRedigoClient(conn)

	tests := []struct {
		name     string
		rh       *rejson.Handler
		days     int
		prices   []*api.DailyPrice
		avgClose float64
		err      string
	}{
		{
			name: "Successfully get price info for the past NDAYS",
			rh:   reJsonHandler,
			days: 3,
			prices: []*api.DailyPrice{
				{
					Day: "2022-04-01",
					Price: &api.Price{
						Open:   "309.3700",
						High:   "310.1300",
						Low:    "305.5400",
						Close:  "309.4200",
						Volume: "27110529",
					},
				},
				{
					Day: "2022-03-31",
					Price: &api.Price{
						Open:   "313.9000",
						High:   "315.1400",
						Low:    "307.8900",
						Close:  "308.3100",
						Volume: "33422070",
					},
				},
				{
					Day: "2022-03-30",
					Price: &api.Price{
						Open:   "313.7600",
						High:   "315.9500",
						Low:    "311.5800",
						Close:  "313.8600",
						Volume: "28163555",
					},
				},
			},
			avgClose: 310.53,
			err:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &storage.Redis{
				Rh: tt.rh,
			}

			// insert price data for the getPrice data to retrieve
			for _, price := range tt.prices {
				res, err := r.Rh.JSONSet(price.Day, ".", price.Price)
				if err != nil {
					t.Errorf("set price :%e", err)
					t.FailNow()
				}

				if res.(string) == "OK" {
					log.Debug().Msg("success")
				} else {
					log.Error().Err(err).Msg("add stock prices")
				}
			}

			prices, avgClose, err := r.GetPriceInfo(tt.days)
			if err != nil {
				assert.Equal(t, err.Error(), tt.err)

				return
			}

			assert.Equal(t, tt.avgClose, avgClose)

			assert.Equal(t, tt.prices, prices)
		})
	}
}

func getSpecificPrice(t *testing.T, rh *rejson.Handler, key string) *api.Price {
	value, err := redis.Bytes(rh.JSONGet(key, "."))
	if err != nil {
		log.Error().Err(err).Str("date", key)
	}

	if value == nil { // no data for that day
		t.Error("nil value")
		t.FailNow()
	}

	price := api.Price{}
	if err = json.Unmarshal(value, &price); err != nil {
		t.Errorf("unmarshal price :%e", err)
		t.FailNow()
	}

	return &price

}
