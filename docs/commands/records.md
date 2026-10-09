# Record commands

Cases are numbered per server, starting at 1, with no gaps. Cases are never deleted. Only the reason can change, and every change is kept.

| Command | Options | Permission |
| --- | --- | --- |
| `/case view` | `case` | Timeout Members |
| `/case reason` | `case`, `reason` | Timeout Members |
| `/history` | `user`, `page` | Timeout Members |
| `/warnings` | `user` | Timeout Members |

- `/case view` shows the case, plus how many times its reason was edited and by whom last.
- `/case reason` replaces the reason, records the old one in the edit history and posts the change to the modlog.
- `/history` lists every case for a user, newest first, 10 per page.
- `/warnings` lists a member's 10 newest warnings and the total count.
