# Automod

Configure with `/config automod`. Every rule is off until turned on. Members with Manage Messages, administrators and the owner are never checked. Bots are ignored.

When a rule fires, the message is deleted. With `/config automod timeout` set, the member is also timed out for that long through the normal moderation path, which creates a case and posts it. Without a timeout, the modlog gets a "Message removed" entry, at most once per member per minute so a flood does not flood the modlog too.

## Message spam

`/config automod spam enabled messages seconds`. Fires when a member sends `messages` messages within `seconds`. Defaults: 6 messages in 5 seconds. Every message past the limit is deleted while the burst lasts.

## Duplicate messages

`/config automod duplicates enabled count seconds`. Fires when the same text is sent `count` times within `seconds`. Case and spacing are ignored. Messages with no text never count. Defaults: 3 times in 30 seconds. Needs the Message Content intent (`MESSAGE_CONTENT=true`).

## Link allowlist

`/config automod links enabled allow`. Fires when a message has an http or https link whose host is not on the list. A domain also allows its subdomains: `youtube.com` allows `www.youtube.com`. Tricks like `https://youtube.com@evil.example` or `https://youtube.com.evil.example` are caught because the real host is checked. Up to 50 domains. Internationalized domains must be written in punycode. Needs the Message Content intent.

## Mention spam

`/config automod mentions limit`. Uses a Discord AutoMod rule named "Keroku mention spam" that blocks messages with more than `limit` unique mentions. `0` removes the rule.

## Invite links

`/config automod invites enabled`. Uses a Discord AutoMod rule named "Keroku invite links" that blocks `discord.gg`, `discord.com/invite` and `dsc.gg` links.

The two Discord AutoMod rules need Keroku to have Manage Server. Keroku only edits rules it created and never touches others.

## What is kept in memory

Per member: the times of recent messages (at most 50) and hashes of recent message text (at most 20), forgotten after 10 minutes of quiet. At most 200,000 members are tracked at once; the least recently active are dropped first. Message text itself is not kept.
