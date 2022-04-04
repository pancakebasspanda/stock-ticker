package api

type JSONResponse struct {
	MetaData    MD              `json:"Meta Data"`
	DailyPrices TimeSeriesDaily `json:"Time Series (Daily)"`
}

type MD struct {
	Information   string `json:"1. Information,omitempty"`
	Symbol        string `json:"2. Symbol,omitempty"`
	LastRefreshed string `json:"3. Last Refreshed,omitempty"`
	OutputSize    string `json:"4. Output Size,omitempty"`
	TimeZone      string `json:"5. Time Zone,omitempty"`
}

type Price struct {
	Open   string `json:"1. open,omitempty"`
	High   string `json:"2. high,omitempty"`
	Low    string `json:"3. low,omitempty"`
	Close  string `json:"4. close,omitempty"`
	Volume string `json:"5. volume,omitempty"`
}

type TimeSeriesDaily map[string]Price

type DailyPrice struct {
	Day   string `json:"Day,omitempty"`
	Price *Price `json:"Time Series (Daily),omitempty"`
}

type OrderedResponse struct {
	DailyPrices     []*DailyPrice `json:"Daily Price,omitempty"`
	AvgClosingPrice float64       `json:"Average Closing Price,omitempty"`
}
