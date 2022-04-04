package api

import (
	"time"
)

type options struct {
	baseURL    string
	timeout    time.Duration
	maxRetries int
	symbol     string
	nDays      int
	apiKey     string
}

// WithKey sets symbol in get request
func WithKey(s string) Option {
	return func(a API) {
		a.(*Client).options.apiKey = s
	}
}

// WithDays num of days data client should retrieve
func WithDays(d int) Option {
	return func(a API) {
		a.(*Client).options.nDays = d
	}
}

// WithSymbol sets symbol in get request
func WithSymbol(s string) Option {
	return func(a API) {
		a.(*Client).options.symbol = s
	}
}

// WithBaseURL sets base URL path for requests
func WithBaseURL(url string) Option {
	return func(a API) {
		a.(*Client).options.baseURL = url
	}
}

// WithTimeout sets a timeout for the http client
func WithTimeout(to time.Duration) Option {
	return func(a API) {
		a.(*Client).options.timeout = to
	}
}

// WithMaxRetries sets times http client will retry the request
func WithMaxRetries(retries int) Option {
	return func(a API) {
		a.(*Client).options.maxRetries = retries
	}
}
