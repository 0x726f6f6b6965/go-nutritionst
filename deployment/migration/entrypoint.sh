#!/bin/bash
set -e

# If CONNECTION_NAME is provided, start the Cloud SQL Auth Proxy in the background
if [ -n "$CONNECTION_NAME" ]; then
  echo "Starting Cloud SQL Auth Proxy for $CONNECTION_NAME..."
  
  # Run the proxy in the background listening on 127.0.0.1:5432
  # It automatically uses active gcloud credentials from /root/.config/gcloud
  PROXY_ARGS="--address 127.0.0.1 --port 5432"
  if [ "$PRIVATE_IP" = "true" ]; then
    echo "Using private IP connection..."
    PROXY_ARGS="$PROXY_ARGS --private-ip"
  fi
  cloud-sql-proxy $PROXY_ARGS "$CONNECTION_NAME" &
  
  # Give the proxy a few seconds to initialize
  echo "Waiting for proxy to start..."
  sleep 3
fi

# Run golang-migrate CLI with all passed arguments
echo "Running migrations..."
exec migrate "$@"
