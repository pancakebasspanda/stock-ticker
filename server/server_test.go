package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"stock_ticker/api"
	"stock_ticker/api/mocks"
	"stock_ticker/storage/mocks"
)

func Test_handler_ServeHTTP(t *testing.T) {
	mockController := gomock.NewController(t)

	storageMock := mock_storage.NewMockStorage(mockController)

	apiMock := mock_api.NewMockAPI(mockController)

	defer mockController.Finish()

	days := 3

	stringResponse := `{"Daily Price":[{"Day":"2022-04-01","Time Series (Daily)":{"1. open":"309.3700","2. high":"310.1300","3. low":"305.5400","4. close":"309.4200","5. volume":"27110529"}},{"Day":"2022-03-31","Time Series (Daily)":{"1. open":"313.9000","2. high":"315.1400","3. low":"307.8900","4. close":"308.3100","5. volume":"33422070"}},{"Day":"2022-03-30","Time Series (Daily)":{"1. open":"313.7600","2. high":"315.9500","3. low":"311.5800","4. close":"313.8600","5. volume":"28163555"}}],"Average Closing Price":313.21}`

	var apiResponse api.OrderedResponse
	if err := json.Unmarshal([]byte(stringResponse), &apiResponse); err != nil {
		t.Errorf("test error :%e", err)
	}

	StockPrices := []*api.DailyPrice{
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
	}

	tests := []struct {
		name                string
		w                   *httptest.ResponseRecorder
		r                   *http.Request
		storageMockOutcomes func(storageMock *mock_storage.MockStorage)
		apiMockOutcomes     func(storageMock *mock_api.MockAPI)
		expected            string
	}{
		{
			name: "successfully get stock price data from cached data",
			w:    httptest.NewRecorder(),
			r:    httptest.NewRequest(http.MethodGet, "/", nil),
			storageMockOutcomes: func(storageMock *mock_storage.MockStorage) {
				storageMock.EXPECT().
					GetPriceInfo(days).
					Times(1).
					Return(StockPrices, 313.21, nil)
			},
			apiMockOutcomes: func(apiMock *mock_api.MockAPI) {
			},
			expected: stringResponse,
		},
		{
			name: "no stock prices in cache so calls the api",
			w:    httptest.NewRecorder(),
			r:    httptest.NewRequest(http.MethodGet, "/", nil),
			storageMockOutcomes: func(storageMock *mock_storage.MockStorage) {
				storageMock.EXPECT().
					GetPriceInfo(days).
					Times(1).
					Return([]*api.DailyPrice{}, float64(0), nil)
			},
			apiMockOutcomes: func(apiMock *mock_api.MockAPI) {
				apiMock.EXPECT().
					GetPrices(gomock.Any()). // contexts will be different each time
					Times(1).
					Return(&apiResponse, nil)
			},
			expected: stringResponse,
		},
		{
			name: "returns an error",
			w:    httptest.NewRecorder(),
			r:    httptest.NewRequest(http.MethodGet, "/", nil),
			storageMockOutcomes: func(storageMock *mock_storage.MockStorage) {
				storageMock.EXPECT().
					GetPriceInfo(days).
					Times(1).
					Return([]*api.DailyPrice{}, float64(0), nil)
			},
			apiMockOutcomes: func(apiMock *mock_api.MockAPI) {
				apiMock.EXPECT().
					GetPrices(gomock.Any()). // contexts will be different each time
					Times(1).
					Return(nil, errors.New("test error"))
			},
			expected: fmt.Sprintf("error: %v", _errResponse),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.storageMockOutcomes(storageMock)
			tt.apiMockOutcomes(apiMock)

			h := &handler{
				apiClient: apiMock,
				redis:     storageMock,
				nDays:     days,
			}

			h.ServeHTTP(tt.w, tt.r)

			res := tt.w.Result()
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			assert.Equal(t, tt.expected, string(data))
		})
	}
}
