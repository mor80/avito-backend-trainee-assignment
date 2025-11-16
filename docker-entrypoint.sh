#!/bin/sh
set -e

if [ -x /app/goose ] && [ -n "$DB_DSN" ]; then
  echo "Waiting for database..."
  until /app/goose -dir ./migrations postgres "$DB_DSN" status >/dev/null 2>&1; do
    sleep 2
  done
  /app/goose -dir ./migrations postgres "$DB_DSN" up
fi

exec "$@"
