# Introduction

Looks like an InfluxDB instance, but writes to a PostgreSQL.

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
    ./mimfluxdb
    ```

1. Make a query:

    ```sh
    echo 'temperature,location=point-of-measurement temperature=64 1744806632000000000'|\
        http POST localhost:8086/api/v2/write 'Authorization: Token some-very-secret-token-here'
    ```

1. Check the app's log for correct parsing of the parts.
