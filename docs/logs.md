# Logs

Set a log channel with `/config logs channel`. Without one, nothing below is posted and no message text is kept.

| Event | Post |
| --- | --- |
| Message edited | Author, channel, text before and after, link to the message. Link preview updates are skipped. |
| Message deleted | Author and text when Keroku saw the message, otherwise "Not cached." |
| Member joined | Member and account age. |
| Member left | Member. Kicks and bans show here too; their cases are in the modlog. |

Bulk deletes from `/purge` are not logged here; the purge itself is in the modlog. Messages from bots are not logged.

Message text needs the Message Content intent (`MESSAGE_CONTENT=true`). Keroku keeps the text of recent messages in memory only, for up to an hour, at most 500 characters each and 50,000 messages in total. It is never written to Postgres.

Posts are queued. During a flood the queue drops posts instead of slowing the bot down.
