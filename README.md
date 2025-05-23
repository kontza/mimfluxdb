# Introduction

Looks like an InfluxDB instance, but writes to a PostgreSQL instance.

## How to Run

1. Create a config file, `$HOME/.config/mimfluxdb/config.toml`:

    ```toml
    tokens = ['some-very-secret-token-here']
    ```

1. Build the project:

    ```sh
    make build
    ```

1. Run the app:

    ```sh
    /tmp/bin/mimfluxdb
    ```

1. Make a query:

    ```sh
    printf "%s,%s %s %s %s %s %s000000000\n" \
        atmospi \
        "location=point-of-measurement" \
        "temperature=6.4" \
        "rssi=-42" \
        "count=16384" \
        "device=28ff248274160427" \
        (date +%s) |\
        http POST localhost:8086/api/v2/write \
        'Authorization: Token some-very-secret-token-here'
    ```

1. Check the app's log for correct parsing of the parts.

## SQL for Grafana

According to <https://duck.ai>, this is the method to get the latest temperature
from each sensor:

```sql
WITH latest AS (
  SELECT device_id,
         MAX(recorded_at) AS latest_recorded_at
  FROM temperature
  GROUP BY device_id
)
SELECT d.label,
       t.recorded_at,
       t.value
FROM temperature t
JOIN latest l
  ON t.device_id = l.device_id
 AND t.recorded_at = l.latest_recorded_at
JOIN device d
  ON t.device_id = d.id;
```

## GORM

1. Install _liquibase_.
1. Run migrations:

    ```sh
    make migrate

    ```

1. Install _gentool_:

    ```sh
    go install gorm.io/gen/tools/gentool@latest
    ```

1. Run it:

    ```sh
    make generate
    ```
