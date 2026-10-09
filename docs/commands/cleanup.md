# Cleanup commands

| Command | Options | Your permission | Keroku needs |
| --- | --- | --- | --- |
| `/purge` | `count`, `user`, `reason` | Manage Messages | Manage Messages, Read Message History |
| `/slowmode` | `interval`, `channel`, `reason` | Manage Channels | Manage Channels |
| `/lock` | `channel`, `duration`, `reason` | Manage Channels | Manage Roles |
| `/unlock` | `channel`, `reason` | Manage Channels | Manage Roles |
| `/lockdown start` | `duration`, `reason` | Manage Server | Manage Roles |
| `/lockdown end` | `reason` | Manage Server | Manage Roles |

Each one posts an entry to the modlog channel. None of them create cases.

## Purge

Deletes up to 500 messages from the channel the command runs in, newest first. With `user`, only that user's messages count. Pinned messages are kept. Discord does not bulk delete messages older than 14 days, so purge stops there and says so.

## Slowmode

`interval` is like `10s` or `5m`, at most `6h`. `off` turns slowmode off. Without `channel`, the current channel changes.

## Lock and unlock

A lock denies @everyone Send Messages, Send Messages in Threads, Create Public Threads, Create Private Threads and Add Reactions in the channel. Keroku remembers those bits from before the lock and puts exactly them back on unlock. Edits to other permissions made while locked are kept. With `duration`, up to 30 days, the channel unlocks on its own.

## Lockdown

A lockdown removes the same permissions from the @everyone role server wide. Channels with an overwrite that allows sending stay open. Ending the lockdown restores the bits @everyone had before. With `duration`, up to 30 days, it ends on its own.
