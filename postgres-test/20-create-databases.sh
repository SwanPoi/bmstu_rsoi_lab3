#!/usr/bin/env bash
set -e
echo "Start creating db"
export VARIANT="v3"
export SCRIPT_PATH=/docker-entrypoint-initdb.d/
export PGPASSWORD=postgres

psql -f "$SCRIPT_PATH/script/db-$VARIANT.sql"
