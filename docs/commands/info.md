# Info commands

Lookups anyone who can see the channel can run. Replies are ephemeral. Nothing is stored.

| Command | Options | Permission |
| --- | --- | --- |
| `/serverinfo` | none | View Channel |
| `/botinfo` | none | View Channel |
| `/userinfo` | `user` | View Channel |
| `/roleinfo` | `role` | View Channel |

- `/serverinfo` shows the owner, creation date, approximate member and online counts, role and emoji counts, boost level and boosts, and verification level.
- `/botinfo` shows the running version, uptime, the approximate number of servers Keroku is in, and the round trip time of one Discord API call.
- `/userinfo` shows the account and when it was created. For members it adds nickname, join date, active timeout and roles. Without `user` it shows you.
- `/roleinfo` shows the role's settings and creation date. Key permissions lists the moderation permissions it grants.

`/userinfo` and `/roleinfo` read the data Discord sends with the command and make no API calls. Mentions in replies never ping.
