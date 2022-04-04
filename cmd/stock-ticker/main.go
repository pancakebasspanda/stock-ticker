package main

import (
	"context"
	"flag"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"net/http"
	"os"
	"os/signal"
	"stock_ticker/storage"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"stock_ticker/api"
	"stock_ticker/server"
)

const (
	_errRedisClient = "redis client initialization error"

	_appName = "stock-ticker"
)

var (
	symbol             string
	nDays              int
	apiKey             string
	maxRetries         int
	baseURL            string
	timeout            int64
	redisURL, redisPWD string
)

func init() {
	flag.StringVar(&baseURL, "base-url", "https://www.alphavantage.co/query", "base url to get stock prices")
	flag.IntVar(&maxRetries, "retries", 3, "max retries")
	flag.Int64Var(&timeout, "timeout", 60, "time in seconds")
}

func main() {
	ctx := context.Background()

	flag.Parse()

	parseEnVars()

	apiClient := api.New(api.WithMaxRetries(maxRetries),
		api.WithBaseURL(baseURL),
		api.WithTimeout(time.Duration(timeout)*time.Second),
		api.WithKey(apiKey),
		api.WithDays(nDays),
		api.WithSymbol(symbol))

	redisClient, err := storage.New(redisURL, redisPWD)
	if err != nil {
		log.Error().Err(err).Msg(_errRedisClient)
	}

	// store full full-length time series of 20+ years in case of rate-limits and NDAYS > 100

	handler := server.NewHandler(apiClient, redisClient, nDays)

	if err = handler.CacheData(ctx); err != nil {
		log.Error().Err(err).Msg("cache data")
	}

	// logger to provide us with free sever metrics
	c := setUpLogger()

	// Here is your final handler as we chain middleware wirth the loggers handler and the servers handler
	h := c.Then(http.HandlerFunc(handler.ServeHTTP))

	mux := http.NewServeMux()

	mux.Handle("/", h) // TODO need better routing

	server := &http.Server{
		Handler:      mux,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// Start Server
	go func() {
		log.Info().
			Str("app", _appName).
			Msg("staring server")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(server)

}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Info().
		Str("app", _appName).
		Msg("shutting down")

	os.Exit(0)
}

func parseEnVars() {
	symbol = getEnv("SYMBOL", "MSFT")

	days := getEnv("NDAYS", "10")

	var err error
	nDays, err = strconv.Atoi(days)
	if err != nil {
		log.Panic().Err(err)
	}

	apiKey = getEnv("API_KEY", "C227WD9W3LUVKVV9")

	redisURL = getEnv("REDIS_URL", "localhost:6379")

	redisPWD = getEnv("REDIS_PASSWORD", "")

}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func setUpLogger() alice.Chain {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("role", "stock-ticker").
		Int("nDays", nDays).
		Str("symbol", symbol).
		Logger()

	c := alice.New()

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(log))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to that handler, all our logs will come with some prepopulated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	return c
}
