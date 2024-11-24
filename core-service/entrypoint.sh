#!/bin/bash

set -e

# Wait for PostgreSQL to be ready
/wait-for-it.sh postgres:5432 -- echo "PostgreSQL is up and running."

# Copy .pgpass file for secure password handling
cp /app/.pgpass /root/.pgpass
chmod 600 /root/.pgpass

# Run migrations
for file in /app/migrations/*.up.sql; do
    echo "Running migration: $file"
    psql -h postgres -U user -d mydb -f "$file" || echo "Migration failed for $file"
done

# Start the core service
./core-service
