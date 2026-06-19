# Copy migrate from official golang-migrate image
FROM migrate/migrate:v4.18.3 AS migrate-cli

# Copy cloud-sql-proxy from official Google image
FROM gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.14.2 AS sql-proxy

# Runner image
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash

# Copy binaries
COPY --from=migrate-cli /migrate /usr/local/bin/migrate
COPY --from=sql-proxy /cloud-sql-proxy /usr/local/bin/cloud-sql-proxy

# Copy migration files into the container
WORKDIR /app
COPY ./migration ./migration

# Copy and setup entrypoint
COPY ./deployment/migration/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
