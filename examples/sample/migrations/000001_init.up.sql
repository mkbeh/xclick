CREATE TABLE IF NOT EXISTS tasks ON CLUSTER '{cluster}'
(
    `id` Int64,
    `description` String,
    `created_at` DateTime
)
    ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}-{uuid}/sample-app/tasks', '{replica}')
    PARTITION BY (id)
    ORDER BY (id)
    SETTINGS index_granularity = 8192;