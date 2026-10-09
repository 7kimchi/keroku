# Settings commands

`/config` needs Manage Server.

| Command | Options |
| --- | --- |
| `/config view` | |
| `/config modlog` | `channel` |
| `/config logs` | `channel` |
| `/config escalation add` | `warnings`, `action`, `duration` |
| `/config escalation remove` | `warnings` |
| `/config automod spam` | `enabled`, `messages`, `seconds` |
| `/config automod duplicates` | `enabled`, `count`, `seconds` |
| `/config automod links` | `enabled`, `allow` |
| `/config automod mentions` | `limit` |
| `/config automod invites` | `enabled` |
| `/config automod timeout` | `duration` |
| `/config raid set` | `enabled`, `joins`, `seconds`, `minage`, `action`, `lockdown` |

## Channels

`/config modlog` sets where cases and cleanup entries go. `/config logs` sets where message edits and deletes and member joins and leaves go. Run either without `channel` to turn it off. Keroku checks it can view, send messages and embed links in the channel before saving.

Changes apply at once on the instance that handled the command and within a minute on other instances.

## Warn escalation

A step runs an action when a member reaches an exact number of warnings. Up to 10 steps per server.

- `warnings`: 1 to 50.
- `action`: `timeout`, `kick` or `ban`.
- `duration`: required for timeouts, optional for bans (a temporary ban), not allowed for kicks.

Escalation actions are made by Keroku, run the same hierarchy checks as manual ones and create their own case. A step fires at most once per warning, even when several warnings land at the same time.

## Automod and raid protection

Each rule is described in [../automod.md](../automod.md). Raid protection is described in [../raid.md](../raid.md).
