# dp-frontend-dataset-controller

==================

An HTTP service for the controlling of data and rendering templates relevant to a particular dataset.

## Configuration

| Environment variable             | Default                          | Description                                                                                                                                           |
| -------------------------------- | -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| API_ROUTER_URL                   | http://localhost:23200/v1        | The URL of the [dp-api-router](https://github.com/ONSdigital/dp-api-router)                                                                           |
| BIND_ADDR                        | :20200                           | The host and port to bind to.                                                                                                                         |
| CACHE_NAVIGATION_UPDATE_INTERVAL | 10s                              | How often the navigation cache is updated                                                                                                             |
| DEBUG                            | false                            | Enable debug mode                                                                                                                                     |
| DOWNLOAD_SERVICE_URL             | http://localhost:23600           | The URL of [dp-download-service](https://www.github.com/ONSdigital/dp-download-service).                                                              |
| ENABLE_MULTIVARIATE              | false                            | Enable 2021 [multivariate datasets](https://github.com/ONSdigital/dp-dataset-api/blob/5f9f4218b65aae4803809f4a876e9f72b9bf5305/models/dataset.go#L43) |
| ENABLE_NEW_NAV_BAR               | false                            | Enable new nav bar                                                                                                                                    |
| ENABLE_PROFILER                  | false                            | Flag to enable go profiler                                                                                                                            |
| GRACEFUL_SHUTDOWN_TIMEOUT        | 5s                               | The graceful shutdown timeout in seconds                                                                                                              |
| HEALTHCHECK_CRITICAL_TIMEOUT     | 90s                              | The time taken for the health changes from warning state to critical due to subsystem check failures                                                  |
| HEALTHCHECK_INTERVAL             | 30s                              | The time between calling healthcheck endpoints for check subsystems                                                                                   |
| OTEL_BATCH_TIMEOUT               | 5s                               | Interval between pushes to OT Collector                                                                                                               |
| OTEL_EXPORTER_OTLP_ENDPOINT      | http://localhost:4317            | URL for OpenTelemetry endpoint                                                                                                                        |
| OTEL_SERVICE_NAME                | "dp-frontend-dataset-controller" | Service name to report to telemetry tools                                                                                                             |
| OTEL_ENABLED                     | false                            | Feature flag to enable OpenTelemetry
| PATTERN_LIBRARY_ASSETS_PATH      | ""                               | Pattern library location                                                                                                                              |
| PPROF_TOKEN                      | ""                               | The profiling token to access service profiling                                                                                                       |
| SITE_DOMAIN                      | localhost                        |                                                                                                                                                       |
| SUPPORTED_LANGUAGES              | []string{"en", "cy"}             | Supported languages                                                                                                                                   |

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

Copyright © 2023, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
