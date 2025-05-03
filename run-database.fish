set PGPASS some-very-secret-password-here
# podman run \
#     --name postgresql \
#     --replace \
#     --rm \
#     --network host \
#     -d \
#     --user 0 \
#     -v ./pg-data/:/bitnami/postgresql:Z \
#     -v ./pg-init/:/docker-entrypoint-initdb.d:Z \
#     --env POSTGRESQL_PASSWORD="$PGPASS" \
#     --privileged \
#     bitnami/postgresql:latest
podman run \
    --name timescaledb \
    --replace \
    --rm \
    --publish 5432:5432 \
    --env POSTGRES_PASSWORD="$PGPASS" \
    --volume ./pg-data/:/home/postgres/pgdata/data:Z \
    timescale/timescaledb-ha:pg17
