package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"net/http"
	"stock_ticker/api"
	"stock_ticker/storage"
	"time"
)

const (
	_errCache    = "error initializes daily prices cache"
	_errResponse = "error retrieving stock price data"
)

type handler struct {
	apiClient api.API
	redis     storage.Storage
	nDays     int
}

func NewHandler(client api.API, redisClient storage.Storage, days int) handler {
	return handler{
		apiClient: client,
		redis:     redisClient,
		nDays:     days,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hlog.FromRequest(r).Info().
		Str("status", "ok").
		Msg("request received")

	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet:
		h.Get(w)
		return
	}

}

// Get is a handler responsible for retrieving the last NDAYS of data
func (h *handler) Get(w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	//TODO check if NDAYS greater than 20 years of stock prices

	var dailyPrices *api.OrderedResponse

	// try retrieving data from the cache
	prices, avgClose, err := h.redis.GetPriceInfo(h.nDays)
	if err != nil {
		log.Error().Err(err).Msg("get cached prices")
	}

	// check if data is in cache
	if len(prices) != 0 && avgClose != 0 {
		dailyPrices = &api.OrderedResponse{
			DailyPrices:     prices,
			AvgClosingPrice: avgClose,
		}
	} else { // call the api to get the data as a fallback
		dailyPrices, err = h.apiClient.GetPrices(ctx)
		if err != nil {
			log.Error().Err(err).Msg("get apiClient prices")
			w.WriteHeader(http.StatusInternalServerError)

			w.Write([]byte(fmt.Sprintf("error: %v", _errResponse)))

			return

		}
	}

	resp, err := json.Marshal(dailyPrices)
	if err != nil {
		log.Error().Err(err).Msg("get apiClient prices")

		w.WriteHeader(http.StatusInternalServerError)

		w.Write([]byte(fmt.Sprintf("error: %v", _errResponse)))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

// CacheData TODO: the api returns the field \Last Refreshed\, this should be checked before caching the data
// CacheData data calls the getFullPriceHistory and stores data to redis to be used when a requests ask for data
func (h *handler) CacheData(ctx context.Context) error {
	resp, err := h.apiClient.GetAllPrices(ctx)
	if err != nil {
		return fmt.Errorf(_errCache+"%e", err)
	}

	if err = h.redis.AddPrices(resp); err != nil {
		return fmt.Errorf(_errCache+"%e", err)
	}

	return err
}
