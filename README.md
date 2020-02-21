dp-frontend-dataset-controller
==================

An HTTP service for the controlling of data relevant to a particular dataset.

### Configuration

| Environment variable         | Default                 | Description
| -----------------------------| ----------------------- | --------------------------------------
| BIND_ADDR                    | :20200                  | The host and port to bind to.
| RENDERER_URL                 | http://localhost:20010  | The URL of [dp-frontend-renderer](https://www.github.com/ONSdigital/dp-frontend-renderer).
| DATASET_API_URL              | http://localhost:22000  | The URL of [dp-dataset-api](https://www.github.com/ONSdigital/dp-dataset-api).
| FILTER_API_URL               | http://localhost:22100  | The URL of [dp-filter-api](https://www.github.com/ONSdigital/dp-filter-api).
| ZEBEDEE_URL                  | http://localhost:8082   | The URL of [zebedee](https://www.github.com/ONSdigital/zebedee).
| DOWNLOAD_SERVICE_URL         | http://localhost:23600  | The URL of [dp-download-service](https://www.github.com/ONSdigital/dp-download-service).
| MAIL_HOST                    | ""                      | The host for the mail server.
| MAIL_PORT                    | ""                      | The port for the mail server.
| MAIL_USER                    | ""                      | A user on the mail server.
| MAIL_PASSWORD                | ""                      | The password for the mail server user.
| FEEDBACK_TO                  | ""                      | Receiver email address for feedback.
| FEEDBACK_FROM                | ""                      | Sender email address for feedback.
| ENABLE_COOKIES_CONTROL       | false                   | Enabled user cookie preferences
| HEALTHCHECK_INTERVAL         | 10s                     | The time between calling healthcheck endpoints for check subsystems
| HEALTHCHECK_CRITICAL_TIMEOUT | 60s                     | The time taken for the health changes from warning state to critical due to subsystem check failures

### Feedback service

The feedback endpoints in this service rely on access to an SMTP service to process feedback.
In order to submit feedback locally, you must have an SMTP service running (try [MailHog](https://www.github.com/mailhog/MailHog))
and update the `MAIL_*` and `FEEDBACK_*` variables configured for this app.

### Licence

Copyright ©‎ 2017 2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
