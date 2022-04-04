package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

const (
	Format = "2006-01-02"
)

// API is an interface to be implemented by the client that connects to it to interact with stock prices API
type API interface {
	GetPrices(ctx context.Context) (*OrderedResponse, error)
	GetAllPrices(ctx context.Context) (*JSONResponse, error)
}

// Option specifies a builder function for configuring a API's client
type Option func(API)

type Client struct {
	options    options
	httpClient *retryablehttp.Client
}

// this is a check to confirm the implementation is compatible with dependent interfaces
var _ API = (*Client)(nil)

// New initializes the api's client
func New(opts ...Option) API {
	client := &Client{}

	for _, opt := range opts {
		opt(client)
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = client.options.maxRetries
	retryClient.HTTPClient.Timeout = client.options.timeout
	retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		// too many requests
		if resp != nil {
			if resp.StatusCode == http.StatusTooManyRequests { // TODO differentiate between this and total number of requests a day
				return 1 * time.Minute // 5 request quota per minute
			}
		}

		// any other error we perform an exponential backoff
		backOff := math.Pow(2, float64(attemptNum)) * float64(min)
		sleep := time.Duration(backOff)
		if float64(sleep) != backOff || sleep > max {
			sleep = max
		}

		return sleep
	}

	client.httpClient = retryClient

	return client
}

func (c *Client) performRequest(ctx context.Context, requestURL string) (io.ReadCloser, error) {
	req, err := c.prepareGetRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	req.Header = map[string][]string{"content-type": {"application/json"},
		"content-encoding": {"gzip"}}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing list stock prices request: %s", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
		return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// GetPrices gets the stock prices for the last NDAYS
// By default, outputsize=compact. The "compact" option is recommended
// if you would like to reduce the data size of each API call as it retrieves 100 days of data
func (c *Client) GetPrices(ctx context.Context) (*OrderedResponse, error) {
	requestURL := fmt.Sprintf("%s?apikey=%s&function=TIME_SERIES_DAILY&symbol=%s", c.options.baseURL, c.options.apiKey, c.options.symbol)

	respBody, err := c.performRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = respBody.Close()
	}()

	prices, avgClose, err := sanitize(respBody, c.options.nDays)
	if err != nil {
		return nil, err
	}

	return &OrderedResponse{
		DailyPrices:     prices,
		AvgClosingPrice: avgClose,
	}, nil

}

// GetAllPrices gets the stock prices for the full-length time series of 20+ years of stock prices
func (c *Client) GetAllPrices(ctx context.Context) (*JSONResponse, error) {
	requestURL := fmt.Sprintf("%s?apikey=%s&function=TIME_SERIES_DAILY&symbol=%s&datatype=json&outputsize=full", c.options.baseURL, c.options.apiKey, c.options.symbol)

	respBody, err := c.performRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = respBody.Close()
	}()

	body, err := ioutil.ReadAll(respBody)

	var res JSONResponse

	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Error().Err(err).Msg("reading response")
	}

	return &res, nil
}

// sanitize returns ndays worth of price data as well as the average closing price
func sanitize(bodyReader io.ReadCloser, nDays int) ([]*DailyPrice, float64, error) {
	// create the response to return
	nDaysData := make([]*DailyPrice, 0)

	body, err := ioutil.ReadAll(bodyReader)

	var res JSONResponse

	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Error().Err(err).Msg("reading response")
	}

	// need to get the contents of NDays and we know its a key value pair in the response
	days := make([]time.Time, len(res.DailyPrices))
	for day := range res.DailyPrices {
		// parse the day into time
		t, err := time.Parse(Format, day)
		if err != nil {
			log.Error().Err(err).Msg("converting daily time series")
		}

		days = append(days, t)
	}

	// sort the time
	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})

	var counter int
	var totClose float64

	for i := len(days) - 1; i >= 0; i-- {
		if counter == nDays {
			break

		}

		var price Price
		var ok bool
		if price, ok = res.DailyPrices[days[i].Format(Format)]; !ok {
			continue
		}

		closePrice, err := strconv.ParseFloat(price.Close, 64)
		if err != nil {
			log.Error().Err(err).Msg("converting close price:")

			continue
		}

		totClose += closePrice

		nDaysData = append(nDaysData, &DailyPrice{
			Day:   days[i].Format(Format),
			Price: &price,
		})

		counter++
	}

	// average close price rounded to 2 decimal places
	avgClose := FixedPrecision(totClose/float64(nDays), 2)

	return nDaysData, avgClose, nil
}

// prepareGetRequest helper function to define the get http request
func (c *Client) prepareGetRequest(ctx context.Context, requestURL string) (*retryablehttp.Request, error) {
	req, err := retryablehttp.NewRequest("GET", requestURL, nil)

	req = req.WithContext(ctx)

	if err != nil {
		return nil, err
	}

	return req, nil
}

// FixedPrecision returns the closing price avg with simple precision formula below. Will fail if numbers flow over float64
func FixedPrecision(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}
