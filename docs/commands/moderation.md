# Moderation commands

Every action runs the same checks, in this order: you hold the required permission, the target is not the server owner, not you, not Keroku, your top role is above the target's (the owner skips this), Keroku's top role is above the target's, and Keroku holds the permission. Roles are fetched from Discord at the moment of the action, never from a cache.

Each action creates a numbered case, posts it to the modlog channel and tries to DM the member. A closed DM never stops the action; the reply shows whether the DM was delivered. Ban and kick DMs go out just before the action, because Keroku cannot reach the member afterwards.

The same request submitted twice, or the same action by two moderators at the same moment, results in one action and one case. The second gets "No change" with the existing case number.

| Command | Options | Permission |
| --- | --- | --- |
| `/ban` | `user`, `reason`, `duration`, `delete` | Ban Members |
| `/unban` | `user`, `reason` | Ban Members |
| `/kick` | `user`, `reason` | Kick Members |
| `/timeout` | `user`, `duration`, `reason` | Timeout Members |
| `/untimeout` | `user`, `reason` | Timeout Members |
| `/warn` | `user`, `reason` (required) | Timeout Members |
| `/note` | `user`, `reason` (required) | Timeout Members |

## Options

- `reason`: up to 512 characters. Control characters and invisible formatting characters are removed. It is shown to the member, in the modlog and in Discord's audit log.
- `duration`: units `w`, `d`, `h`, `m`, `s`, largest first, like `1d4h` or `30m`. At least 1 minute.
- `/ban duration`: up to 365 days. Empty means permanent. A temporary ban is lifted automatically and the unban gets its own case.
- `/ban delete`: removes the member's messages from the last hour, 6 hours, day, 3 days or 7 days.
- `/timeout duration`: up to 365 days. Discord caps a timeout at 28 days, so longer timeouts are renewed automatically until they end.

## Notes and warnings

Notes are private. The member is not told and no DM is sent. Warnings DM the member and count toward warn escalation, see [settings.md](settings.md).

## When something fails

If Discord refuses or does not answer, nothing is recorded and the reply says so. If Discord applied the action but the case could not be saved, Keroku reverts the ban, unban or timeout and says it did. Kicks cannot be reverted; the reply says the kick happened without a case and includes a reference id for the logs.
