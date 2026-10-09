# Security

## Threat model

Keroku assumes every interaction, gateway event and option value can be hostile: forged ids, oversized strings, markdown and mention injection, replayed interactions, moderators trying to act above their rank, and floods from many accounts at once. It also assumes Discord and Postgres can fail at any moment.

| Threat | Defense |
| --- | --- |
| Acting above your rank | Every action re-checks permission and role hierarchy in the handler, with members fetched fresh from Discord. |
| Acting on the owner, yourself or Keroku | Refused before any Discord call. |
| Permission overrides | Each command checks its default permission even if a server admin exposed it to other roles. |
| Replayed or double submitted commands | One case per interaction, enforced by a unique constraint. A transaction lock on the target turns concurrent duplicates into "No change". |
| Cross server access | Every query filters on the server id. Tests prove a case from one server is invisible from another. |
| Mention abuse | Every message disables all mentions. User text in embeds is escaped. |
| Spoofed text | Reasons are trimmed and stripped of control, zero width and bidi characters, and capped at 512 characters. |
| Malformed input | Ids, durations and counts are parsed strictly. Bad input is rejected, never adjusted. |
| Floods | Per user and per server command limits, bounded queues that drop instead of growing, and size capped caches. |
| Token leaks | The token is read from the environment or a mounted file and redacted from every log line. Errors shown in Discord carry a reference id, not internals. |
| Crashes | A panic in one event or command is recovered and logged. The process keeps running. |

## What is stored

| Data | Where | Kept |
| --- | --- | --- |
| Cases: server id, case number, action, target id, moderator id, reason, duration, time | Postgres | Forever. Cases are never deleted. |
| Reason edits: editor id, old and new reason, time | Postgres | Forever. |
| Server settings: channel ids, escalation steps, automod and raid settings | Postgres | Until changed. |
| Pending timers: unbans, timeout renewals, unlocks, lockdown ends | Postgres | 30 days after they finish. |
| Channel locks and lockdowns: the permission bits to restore | Postgres | Until unlocked. |
| Interaction ids already handled | Postgres | 24 hours. |
| Rate limit and dedupe state | Memory | Minutes, bounded in size. |

Keroku does not store message content in Postgres. Usernames are not stored; embeds render user ids as mentions.

## Reporting

Report vulnerabilities privately to the maintainer. Do not open a public issue.
