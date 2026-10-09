# Keroku

Discord moderation and logging bot. Slash commands only. Written in Go, backed by PostgreSQL.

Keroku connects to Discord over the gateway and opens no public port. Command replies are ephemeral: only the member who ran the command sees them. Every moderation case is posted to the modlog channel.

## Run

Requirements: Go 1.27.2 or newer, PostgreSQL 18 or newer (the bot refuses to start on older servers), a Discord application with a bot user.

```sh
export DISCORD_TOKEN=...            # or DISCORD_TOKEN_FILE=/run/secrets/token
export DATABASE_URL=postgres://keroku@db/keroku
go run ./cmd/keroku
```

The schema is migrated on startup. Several instances can start at once: migrations take a Postgres advisory lock. Every setting is listed in [docs/config.md](docs/config.md). Deployment, sharding and metrics are in [docs/operations.md](docs/operations.md).

## Commands

[Moderation](docs/commands/moderation.md), [records](docs/commands/records.md), [cleanup](docs/commands/cleanup.md), [settings](docs/commands/settings.md) and [info](docs/commands/info.md).

## Discord permissions

Invite Keroku with exactly these permissions and place its role above every role it should be able to moderate.

| Permission | Used by |
| --- | --- |
| View Channels | Reading channels it posts to or cleans up. |
| Send Messages | Modlog and log posts. |
| Embed Links | Every post is an embed. |
| Read Message History | `/purge`. |
| Manage Messages | `/purge`, automod deletions. |
| Manage Channels | `/slowmode`. |
| Manage Roles | `/lock`, `/unlock`, `/lockdown`. |
| Kick Members | `/kick`, escalation, raid protection. |
| Ban Members | `/ban`, `/unban`, escalation, raid protection. |
| Timeout Members | `/timeout`, `/untimeout`, escalation, automod. |
| Manage Server | Discord AutoMod rules for mention spam and invites. |

## Gateway intents

Enable these in the Discord developer portal. Keroku requests nothing else.

| Intent | Privileged | Used for |
| --- | --- | --- |
| Guilds | no | Guild and channel lifecycle. |
| Guild Members | yes | Raid detection and join and leave logs. |
| Guild Moderation | no | Ban events. |
| Guild Messages | no | Automod and message logs. |
| Message Content | yes | Only when `MESSAGE_CONTENT=true`. Needed by the duplicate message and link rules and by edit and delete logs that show text. |

## Development

```sh
scripts/installTools.sh   # pinned linters
scripts/check.sh          # every gate, including tests against a throwaway Postgres
```

Tests need PostgreSQL 18 binaries on the path. `scripts/testDb.sh start` starts a local test server in `.tmp/`.

## License

Source available. You may run and modify Keroku for servers you own or administer. You may not redistribute it or host it for other servers. See [LICENSE](LICENSE).
