package storage

import (
	"encoding/json"
	"fmt"
	"github.com/nitishm/go-rejson/v4"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"

	"stock_ticker/api"
)

// Storage is the interface for storage operations
type Storage interface {
	AddPrices(prices *api.JSONResponse) error
	GetPriceInfo(days int) ([]*api.DailyPrice, float64, error)
}

// Redis is the implementation of Storage interface
type Redis struct {
	Rh *rejson.Handler
}

func New(address, password string) (Storage, error) {
	reJsonHandler := rejson.NewReJSONHandler()

	// Redigo Client
	conn, err := redis.Dial("tcp", address, redis.DialPassword(password))
	if err != nil {
		return nil, err
	}

	reJsonHandler.SetRedigoClient(conn)

	return &Redis{
		Rh: reJsonHandler,
	}, nil
}

func (r *Redis) AddPrices(prices *api.JSONResponse) error {
	for date, price := range prices.DailyPrices {
		res, err := r.Rh.JSONSet(date, ".", price)

		if err != nil {
			return err
		}

		if res.(string) != "OK" {
			log.Error().Err(err).Msg("add stock prices")
		}
	}

	return nil
}

func (r *Redis) GetPriceInfo(days int) ([]*api.DailyPrice, float64, error) {
	nDaysData := make([]*api.DailyPrice, 0)
	var totClose float64

	var counter int

	// 20 years worth of daily data is the maximum
	maxDays := int(time.Now().Sub(time.Now().AddDate(-20, 0, 0)).Hours() / 24)

	//TODO as an alternate insert prices as one object array and call getObject keys so we dont rely on making the key
	for i := 1; i <= maxDays; i++ {

		if counter == days {
			break
		}

		key := time.Now().AddDate(0, 0, -i).Format(api.Format) // api gets up to yesterdays date of data

		value, err := redis.Bytes(r.Rh.JSONGet(key, "."))
		if err != nil {
			log.Error().Err(err).Str("date", key)
		}

		if value == nil { // no data for that day
			continue
		}

		price := api.Price{}
		if err = json.Unmarshal(value, &price); err != nil {
			log.Error().Err(err).Msg("unmarshal price:")
		}

		closePrice, err := strconv.ParseFloat(price.Close, 64)
		if err != nil {
			log.Error().Err(err).Msg("converting close price:")

			return nil, 0, fmt.Errorf("converting close price: %v", err)
		}

		totClose += closePrice

		nDaysData = append(nDaysData, &api.DailyPrice{
			Day:   key,
			Price: &price,
		})

		counter++
	}

	// average close price rounded to 2 decimal places
	avgClose := api.FixedPrecision(totClose/float64(days), 2)

	return nDaysData, avgClose, nil
}
