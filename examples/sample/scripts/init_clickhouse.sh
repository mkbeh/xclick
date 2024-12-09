#!/bin/bash
set -e

clickhouse client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" -n <<-EOSQL
    CREATE DATABASE IF NOT EXISTS "$CLICKHOUSE_DB";
EOSQL