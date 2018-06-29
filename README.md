dp-frontend-dataset-controller
==================

An HTTP service for the controlling of data relevant to a particular dataset.

### Configuration

| Environment variable | Default                 | Description
| -------------------- | ----------------------- | --------------------------------------
| BIND_ADDR            | :20200                  | The host and port to bind to.
| RENDERER_URL         | http://localhost:20010  | The URL of dp-frontend-renderer.
| DATASET_API_URL      | http://localhost:22000  | The URL of the dp-dataset-api.
| FILTER_API_URL       | http://localhost:22100  | The URL of the dp-filter-api.
| ZEBEDEE_URL          | http://localhost:8082   | The URL of zebedee.
| SLACK_TOKEN          |                         | A slack token to write feedback to slack

### Licence

Copyright ©‎ 2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

