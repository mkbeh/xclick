# ClickHouse Library

This library provides an API for working with ClickHouse, using [clickhouse-go](github.com/ClickHouse/clickhouse-go) and
integration with OpenTelemetry for tracing and metrics.

## Features

- Query Builder such as [squirrel](github.com/Masterminds/squirrel)
- Built-in migrations using [golang-migrate](github.com/golang-migrate/migrate)
- Observability

## Getting started

Here's a basic overview of using (more examples can be found [here](github.com/mkbeh/clickhouse/examples)):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mkbeh/clickhouse"
)

func main() {
	cfg := &clickhouse.Config{
		Hosts:          "127.0.0.1:8123",
		User:           "user",
		Password:       "password",
		DB:             "sample",
		MigrateEnabled: true,
	}

	pool, err := clickhouse.NewPool(
		clickhouse.WithConfig(cfg),
		clickhouse.WithClientID("test-client"),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer pool.Close()

	var greeting string
	err := pool.QueryRow(context.Background(), "select 'hello world'").Scan(&greeting)
	if err != nil {
		log.Fatalln("QueryRow failed", err)
	}

	fmt.Println(greeting)
}

```

## Migrations

Full example can be found [here](github.com/mkbeh/clickhouse/examples).

Create file `embed.go` in your migrations directory:

```go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

Pass `embed.FS` with option `WithMigrations`

```go
pool, _ := clickhouse.NewPool(
...
postgres.WithMigrations(migrations.FS),
)
```

## Configuration

Available client options:

| ENV                                    | Required | Description                                                                                                                                                                                  |
|----------------------------------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| CLICKHOUSE_SHARD_ID                    | -        | Shard ID (default 0).                                                                                                                                                                        |
| CLICKHOUSE_HOSTS                       | true     | Comma-separated list of single address hosts for load-balancing and failover.                                                                                                                |
| CLICKHOUSE_USER                        | true     | Auth credentials.                                                                                                                                                                            |
| CLICKHOUSE_PASSWORD                    | true     | Auth credentials.                                                                                                                                                                            |
| CLICKHOUSE_DB                          | true     | Select the current default database.                                                                                                                                                         |
| CLICKHOUSE_MAX_OPEN_CONNS              | -        | Max open connections (default: 32)                                                                                                                                                           |
| CLICKHOUSE_MAX_IDLE_CONNS              | -        | Max idle connections (default: 8)                                                                                                                                                            |
| CLICKHOUSE_CONN_MAX_LIFETIME           | -        | Connection max lifetime (default: 1h)                                                                                                                                                        |
| CLICKHOUSE_DIAL_TIMEOUT                | -        | A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix such as "300ms", "1s". Valid time units are "ms", "s", "m". (default 30s). |
| CLICKHOUSE_READ_TIMEOUT                | -        | A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix such as "300ms", "1s". Valid time units are "ms", "s", "m" (default 5m).   |
| CLICKHOUSE_DEBUG                       | -        | Enable debug output (boolean value).                                                                                                                                                         |
| CLICKHOUSE_FREE_BUFFER_ON_CONN_RELEASE | -        | Drop preserved memory buffer after each query.                                                                                                                                               |
| CLICKHOUSE_INSECURE_SKIP_VERIFY        | -        | Skip certificate verification (default is false)                                                                                                                                             |
| CLICKHOUSE_BLOCK_BUFFER_SIZE           | -        | Size of block buffer (default 2).                                                                                                                                                            |
| CLICKHOUSE_MAX_COMPRESSION_BUFFER      | -        | Max size (bytes) of compression buffer during column by column compression (default 10MiB)                                                                                                   |
| CLICKHOUSE_HTTP_HEADERS                | -        | Set additional headers on HTTP requests.                                                                                                                                                     |
| CLICKHOUSE_HTTP_URL_PATH               | -        | Set additional URL path for HTTP requests.                                                                                                                                                   |
| CLICKHOUSE_CONN_OPEN_STRATEGY          | -        | Random/round_robin/in_order (default in_order).                                                                                                                                              |
| CLICKHOUSE_SETTINGS                    | -        | ClickHouse settings.                                                                                                                                                                         |
| CLICKHOUSE_MIGRATE_ENABLED             | -        | Enable migrations if passed (default false).                                                                                                                                                 |
| CLICKHOUSE_MIGRATE_ARGS                | -        | Additional arguments for connection string.                                                                                                                                                  |

Additional args that can be added:

* `CLICKHOUSE_MIGRATE_ARGS = x-multi-statement=true&x-cluster-name=distributed_cluster&x-migrations-table-engine=ReplicatedMergeTree`