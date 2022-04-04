# Stock Ticker

**About** <br />
Stock Ticker is a web service that looks up a fixed number of closing prices of a specific stock.
For a given GET request the web service should return the last NDAYS of data along with the average closing price over those days.

For more information about the api used to retrieve the stock prices please see [alphavantage: TIME_SERIES_DAILY](https://www.alphavantage.co/documentation/)

This API returns raw (as-traded) daily time series (date, daily open, daily high, daily low, daily close, daily volume)
of the global equity specified, covering 20+ years of historical data. If you are also interested in split/dividend-adjusted historical data, please use the Daily Adjusted API, which covers adjusted close values and historical split and dividend events.
The above API has the following limitations
* 5 API requests per minute
* 500 requests per day

In order to cater for the above limits the app connects to and stores the raw api data redis cache and specifically uses the redis module [RedisJSON](https://redis.io/docs/stack/json/)
. The choice of RedisJSON was due to it being quick and easy to implement given the time constraint plus the quickness of looking up data.

When the application first starts up, it does a call to retrieve the FULL STOCK PRICE history of specific stock(20 Years). This is then cached in
redis. Thereafter, when a request comes in to retrieve NDAYS worth of data it first looks up dates in the cache and if that fails it will then call the API therefore not using
up API quota on each request. 

The app also uses [zerolog](https://github.com/rs/zerolog) due to its integration with the net/http package where it has helpers to integrate 
zerolog with http.Handler for some request context fields.  In order to do this we needed to chain log handler and the webservices default handler so [alice](https://github.com/justinas/alice) was used

A kubernetes [config map](https://kubernetes.io/docs/concepts/configuration/configmap/) was used to pass in all environment variables and a [kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/) was used to pass in the 

**Assumptions** <br />
* Using a vanilla kubernetes environment
* The config and the app runs for specific stock. If more stocks are required we can have different deployments reading the config map and being a specific deployment per stock or being a multi deployments and 
caching data for all stock
* Running one instance of redis as we are not managing aa connection pool, and we want high availability


**Local Machine Run using Docker containers** <br />
In order to run the stock-ticker service, you first need to run the docker containers
running redis.  This is achieved by running the docker-compose.yaml file as follows

```shell
$ docker-compose up

 - Container stock-ticker-redis-1         Running                                                                                                                                  0.0s 
 - Container stock-ticker-stock-ticker-1  Created                                                                                                                                  0.4s 
Attaching to stock-ticker-redis-1, stock-ticker-stock-ticker-1
stock-ticker-stock-ticker-1  | 2022/04/04 23:04:08 [DEBUG] GET https://www.alphavantage.co/query?apikey=C227WD9W3LUVKVV9&function=TIME_SERIES_DAILY&symbol=MSFT&datatype=json&outputsize
=full
stock-ticker-redis-1         | 1:M 04 Apr 2022 23:04:10.262 * 1 changes in 3600 seconds. Saving...
stock-ticker-redis-1         | 1:M 04 Apr 2022 23:04:10.262 * Background saving started by pid 21
stock-ticker-redis-1         | 21:C 04 Apr 2022 23:04:10.288 * DB saved on disk
stock-ticker-redis-1         | 21:C 04 Apr 2022 23:04:10.289 * RDB: 1 MB of memory used by copy-on-write
stock-ticker-redis-1         | 1:M 04 Apr 2022 23:04:10.362 * Background saving terminated with success
stock-ticker-stock-ticker-1  | {"level":"info","app":"stock-ticker","time":"2022-04-04T23:04:11Z","message":"staring server"}

```


```shell

$ kubectl apply -f namespace.yaml

namespace/stocks created

$ kubectl apply -f redis.yaml 

deployment.apps/redis-master created
service/redis-master created

$ kubectl apply -f secret.yaml
secret/secret created

$ kubectl apply -f config.yaml
configmap/config created

$ kubectl apply -f stock-ticker.yaml
deployment.apps/stock-ticker created
service/stock-ticker created

$ >kubectl logs -f stock-ticker-59f4fcbbc9-vmd6z  -n stocks 
2022/04/04 22:48:36 [DEBUG] GET https://www.alphavantage.co/query?apikey=C227WD9W3LUVKVV9&function=TIME_SERIES_DAILY&symbol=MSFT&datatype=json&outputsize=full
{"level":"info","app":"stock-ticker","time":"2022-04-04T22:48:49Z","message":"staring server"}

```

***Examples*** <br />
GET request for the last NDAYS(10 days) of stock prices and the average closing price:

```shell
wget -O response.json http://localhost:8080/
```
```json
{
  "Daily Price":[
    {
      "Day":"2022-04-01",
      "Time Series (Daily)":{
        "1. open":"309.3700",
        "2. high":"310.1300",
        "3. low":"305.5400",
        "4. close":"309.4200",
        "5. volume":"27110529"
      }
    },
    {
      "Day":"2022-03-31",
      "Time Series (Daily)":{
        "1. open":"313.9000",
        "2. high":"315.1400",
        "3. low":"307.8900",
        "4. close":"308.3100",
        "5. volume":"33422070"
      }
    },
    {
      "Day":"2022-03-30",
      "Time Series (Daily)":{
        "1. open":"313.7600",
        "2. high":"315.9500",
        "3. low":"311.5800",
        "4. close":"313.8600",
        "5. volume":"28163555"
      }
    },
    {
      "Day":"2022-03-29",
      "Time Series (Daily)":{
        "1. open":"313.9100",
        "2. high":"315.8200",
        "3. low":"309.0500",
        "4. close":"315.4100",
        "5. volume":"30393403"
      }
    },
    {
      "Day":"2022-03-28",
      "Time Series (Daily)":{
        "1. open":"304.3300",
        "2. high":"310.8000",
        "3. low":"304.3300",
        "4. close":"310.7000",
        "5. volume":"29578188"
      }
    },
    {
      "Day":"2022-03-25",
      "Time Series (Daily)":{
        "1. open":"305.2300",
        "2. high":"305.5000",
        "3. low":"299.2855",
        "4. close":"303.6800",
        "5. volume":"22443541"
      }
    },
    {
      "Day":"2022-03-24",
      "Time Series (Daily)":{
        "1. open":"299.1400",
        "2. high":"304.2000",
        "3. low":"298.3150",
        "4. close":"304.1000",
        "5. volume":"24484456"
      }
    },
    {
      "Day":"2022-03-23",
      "Time Series (Daily)":{
        "1. open":"300.5100",
        "2. high":"303.2300",
        "3. low":"297.7201",
        "4. close":"299.4900",
        "5. volume":"25715377"
      }
    },
    {
      "Day":"2022-03-22",
      "Time Series (Daily)":{
        "1. open":"299.8000",
        "2. high":"305.0000",
        "3. low":"298.7700",
        "4. close":"304.0600",
        "5. volume":"27441386"
      }
    },
    {
      "Day":"2022-03-21",
      "Time Series (Daily)":{
        "1. open":"298.8900",
        "2. high":"300.1400",
        "3. low":"294.9000",
        "4. close":"299.1600",
        "5. volume":"28107855"
      }
    }
  ],
  "Average Closing Price":306.82
}
```

## Running tests
First we need to start our docker test database in which the tests interact with.
```shell
docker-compose up redis 
```
Then you can run all tests with
```shell
go test $(go list ./... | grep -v /vendor/ | grep -v /cmd/) -race
```

## Upcoming Changes and Features
* ***Add concurrency***<br />
When getting the cache that should be in a goroutine as to not block the starting up of the service. This would include adding a readiness probe.<br />
* ***Health and freshness check on stocks API***<br />
* ***Clean up model package and separation between API model and server model***<br />
* ***Clean up code in regard to TODO's left in the codebase*** <br />
Some examples here include optimizing parameters in functions<br />
* ***Reliability concerns***<br />
* Scaling this service requires additional config to run for all types of stocks
* Internal rate limiting given we have too many requests and resources are limited
* Horizontal pod auto-scalar based on traffic

`
## Packages

`/.kube`: kubernetes manifests that include secrets, config-maps, namespace and deployments

`/cmd`: main.go + setup directly related to logging and reading ENV variables

`/server`: http server implementation

`/api`: interface that is implemented and that interacts with stock prices API

`/storage`: Storage interface (we used redis in this instance as the implementation)

`/integration-tests`: tests that directly test the storage implementation against a test redis db

