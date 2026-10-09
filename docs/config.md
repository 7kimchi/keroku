# Configuration

Keroku reads its settings from environment variables at startup. Invalid values stop the process with a message naming the variable. Nothing is coerced: a bad value is an error, not a default.

## Secrets

Set each secret either directly or as a path to a mounted file. Setting both is an error. A trailing newline in a secret file is ignored.

| Variable | File variant | Value |
| --- | --- | --- |
| `DISCORD_TOKEN` | `DISCORD_TOKEN_FILE` | Raw bot token, without the `Bot ` prefix. |
| `DATABASE_URL` | `DATABASE_URL_FILE` | Postgres connection string. |

Secrets never appear in logs. The logger replaces the token, the database URL and anything shaped like a Discord token with `[redacted]`.

## Runtime

| Variable | Default | Range | Purpose |
| --- | --- | --- | --- |
| `METRICS_ADDR` | `127.0.0.1:9100` | `127.0.0.1:port` or `off` | Prometheus listener. Only loopback is accepted. |
| `SHARD_COUNT` | `0` | 0 to 4096 | Total shards. `0` uses Discord's recommendation. |
| `SHARD_IDS` | all | ids or ranges, like `0-3,8` | Shards this instance runs. Needs `SHARD_COUNT`. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` | Minimum log level. |
| `WORKERS` | `32` | 1 to 1024 | Parallel event workers. |
| `QUEUE_SIZE` | `256` | 1 to 100000 | Queued events per worker before new events are dropped. |
| `DB_MAX_CONNS` | `40` | 2 to 1000 | Postgres pool size. Must be at least `WORKERS` plus 4. |
| `MESSAGE_CONTENT` | `false` | `true` or `false` | Request the message content intent. Needed for the duplicate and link automod rules. |
| `SHUTDOWN_TIMEOUT` | `20s` | 1s to 5m | Time allowed to finish in-flight work on shutdown. |

## Database session limits

Every pooled connection runs with `statement_timeout` 10s, `lock_timeout` 10s and `idle_in_transaction_session_timeout` 60s. Migrations lift the first two on their own connection.
