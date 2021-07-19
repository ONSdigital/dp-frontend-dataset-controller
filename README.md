# dp-frontend-dataset-controller

==================

An HTTP service for the controlling of data relevant to a particular dataset.

## Configuration

| Environment variable         | Default                 | Description
| -----------------------------| ----------------------- | --------------------------------------
| BIND_ADDR                    | :20200                  | The host and port to bind to.
| DEBUG                        | false                   | Enable debug mode
| API_ROUTER_URL               | http://localhost:23200/v1 | The URL of the [dp-api-router](https://github.com/ONSdigital/dp-api-router)
| SITE_DOMAIN                  | localhost               |
| PATTERN_LIBRARY_ASSETS_PATH  | ""                      | Pattern library location
| SUPPORTED_LANGUAGES          | [2]string{"en", "cy"}   | Supported languages
| DOWNLOAD_SERVICE_URL         | http://localhost:23600  | The URL of [dp-download-service](https://www.github.com/ONSdigital/dp-download-service).
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                      | The graceful shutdown timeout in seconds
| HEALTHCHECK_INTERVAL         | 30s                     | The time between calling healthcheck endpoints for check subsystems
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                     | The time taken for the health changes from warning state to critical due to subsystem check failures
| ENABLE_PROFILER              | false                   | Flag to enable go profiler
| PPROF_TOKEN                  | ""                      | The profiling token to access service profiling

## Profiling

An optional `/debug` endpoint has been added, in order to profile this service via `pprof` go library.
In order to use this endpoint, you will need to enable profiler flag and set a PPROF_TOKEN:

```
export ENABLE_PROFILER=true
export PPROF_TOKEN={generated uuid}
```

Then you can us the profiler as follows:

1- Start service, load test or if on environment wait for a number of requests to be made.

2- Send authenticated request and store response in a file (this can be best done in command line like so: `curl <host>:<port>/debug/pprof/heap -H "Authorization: Bearer {generated uuid} > heap.out` - see pprof documentation on other endpoints

3- View profile either using a web ui to navigate data (a) or using pprof on command line to navigate data (b) 
  a) `go tool pprof -http=:8080 heap.out`
  b) `go tool pprof heap.out`, -o flag to see various options

## Licence

Copyright ©‎ 2017 2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
