# Raid protection

Configure with `/config raid set`. Off until `enabled` is true. Options left empty keep their current value.

| Option | Default | Range | Meaning |
| --- | --- | --- | --- |
| `enabled` | false | | Turns raid protection on or off. |
| `joins` | 10 | 2 to 500 | Joins that count as a raid. |
| `seconds` | 10 | 1 to 600 | Window those joins must fall in. |
| `minage` | off | 1m to 365d, or off | Accounts younger than this are removed when they join. |
| `action` | kick | kick, ban, lockdown | What happens to a raid. |
| `lockdown` | 15m | 1m to 7d | How long a raid lockdown lasts. |

## Join bursts

When `joins` members join within `seconds` of each other, Keroku posts "Raid detected" to the modlog and then:

- `kick` or `ban`: removes every member who joined in the window, and every member who joins while the burst continues. A raid ends once a full window passes with fewer joins. Bans also delete the last hour of their messages.
- `lockdown`: starts a server lockdown for the `lockdown` duration. See [commands/cleanup.md](commands/cleanup.md).

Each removal is a case made by Keroku, with the same hierarchy checks as a manual action.

## Account age

With `minage` set, a member whose account is younger is kicked on join, or banned when `action` is `ban`. With `action` set to `lockdown`, young accounts are kicked.

## Limits

Join windows are kept in memory per server, at most 1,000 joins each and 100,000 servers in total. A restart forgets them, which only delays detection by one window.
