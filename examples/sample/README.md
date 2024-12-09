# Description

This is a sample REST API service implemented using clickhouse-go as the connector to a ClickHouse data store.

# Usage

Create a ClickHouse database.

Configure the database connection with environment variables:

```text
CLICKHOUSE_HOSTS=localhost:9001
CLICKHOUSE_USER=sample_app
CLICKHOUSE_PASSWORD=sample_app
CLICKHOUSE_DB=sample_app
CLICKHOUSE_MIGRATE_ARGS=x-multi-statement=true&x-cluster-name=distributed_cluster&x-migrations-table-engine=ReplicatedMergeTree
```

Run main.go:

```
go run main.go
```

## Create tasks

```shell
curl '127.0.0.1:8080/create'
```

## Get tasks

```shell
curl '127.0.0.1:8080/get'
```

## Metrics

```shell
curl 'http://localhost:8080/metrics'
```