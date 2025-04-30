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
    printf "%s,%s %s %s %s %s %s\n" \
        temperatures \
        "location=point-of-measurement" \
        "temperature=6.4" \
        "rssi=-42" \
        "count=16384" \
        "device=28ff248274160427" \
        17448066320000000000 |\
        http POST localhost:8086/api/v2/write \
        'Authorization: Token some-very-secret-token-here'
    ```

1. Check the app's log for correct parsing of the parts.
