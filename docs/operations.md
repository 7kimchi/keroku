# Operations

## Deploy

Run the binary with the variables in [config.md](config.md). It needs outbound HTTPS to Discord and a connection to Postgres. It listens on nothing public. On SIGTERM or SIGINT it stops reading gateway events, finishes queued commands within `SHUTDOWN_TIMEOUT`, then closes the database pool.

## Shards

With `SHARD_COUNT=0` and no `SHARD_IDS`, one instance runs every shard Discord recommends. Shards connect in batches of Discord's `max_concurrency`, 5 seconds apart. Startup fails if the daily session start allowance is too low for the shards requested.

To split shards across instances, give every instance the same `SHARD_COUNT` and a different `SHARD_IDS`:

```sh
# instance a
SHARD_COUNT=8 SHARD_IDS=0-3
# instance b
SHARD_COUNT=8 SHARD_IDS=4-7
```

Only the instance running shard 0 registers slash commands. Command and event handling for a guild happens on the instance that owns the guild's shard. Timers and cross instance work coordinate through Postgres.

## Metrics

Prometheus metrics are served on `METRICS_ADDR`, `127.0.0.1:9100` by default, path `/metrics`. Only loopback is allowed. Scrape it from a sidecar in the same network namespace, or set `METRICS_ADDR=off`.

| Metric | Meaning |
| --- | --- |
| `keroku_events_total{type}` | Gateway events received. |
| `keroku_events_dropped_total{reason}` | Events dropped: duplicates, full queues, expired tokens. |
| `keroku_commands_total{command,result}` | Commands by outcome: `ok`, `refused`, `error`, `panic`. |
| `keroku_command_seconds{command}` | Handler time. |
| `keroku_rate_limited_total{scope}` | Commands refused by the per user or per guild limit. |
| `keroku_discord_errors_total{kind}` | Failed Discord API calls by kind. |
| `keroku_panics_total{where}` | Recovered panics. Any value above 0 is a bug. |
| `keroku_queue_depth{queue}` | Work waiting in the `ack` and `lanes` queues. |

## Backups

All state lives in Postgres. Back it up with `pg_dump` or continuous archiving. Losing the database loses cases, settings and pending timers. Nothing else needs a backup.
