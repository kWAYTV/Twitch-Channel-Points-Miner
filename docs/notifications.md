# Notifications

The miner can send what happens to a Discord channel, a Telegram chat, or both. Each is set up on its own and has its own list of events.

```json
"discord": {
  "webhook": "https://discord.com/api/webhooks/...",
  "events": []
},
"telegram": {
  "token": "123456:ABC...",
  "chat_id": 123456789,
  "events": ["STREAMER_ONLINE", "BET_WIN", "BET_LOSE"],
  "silent": false
}
```

## Discord

1. In Discord, open the channel's settings, then Integrations, then Webhooks, and create a webhook.
2. Copy its URL into `discord.webhook`.

An empty `webhook` turns Discord off.

## Telegram

1. Talk to [@BotFather](https://t.me/BotFather), send `/newbot` and follow the steps. Copy the token it gives you into `telegram.token`.
2. Open a chat with your new bot and send it any message. A bot cannot write to you first.
3. Set `telegram.chat_id` to your numeric chat id. For a public channel the bot is an admin of, `"@channelname"` works too.

`"silent": true` delivers messages silently. An empty `token` or `chat_id` turns Telegram off.

Treat the token and the webhook URL like passwords: anyone who has one can post as your bot or to your channel.

## Choosing events

`events` lists what to send.

- `[]` sends everything except three frequent events: `GAIN_FOR_WATCH`, `DROP_STATUS` and `BET_FILTERS`.
- `["*"]` or `["ALL"]` sends everything, those three included.
- Otherwise only the events you name are sent.

| Event | Sent when |
|---|---|
| `STARTUP` | The session starts, and once the streamers are loaded. |
| `SHUTDOWN` | The session ends. |
| `STREAMER_ONLINE`, `STREAMER_OFFLINE` | A streamer goes live or stops. |
| `GAIN_FOR_WATCH` | Points arrive for watching. Frequent. |
| `GAIN_FOR_WATCH_STREAK` | A watch streak bonus arrives. |
| `GAIN_FOR_CLAIM` | A bonus chest is claimed. |
| `GAIN_FOR_RAID` | Points arrive for joining a raid. |
| `JOIN_RAID` | The miner joins a raid. |
| `MOMENT_CLAIM` | A Moment is claimed. |
| `DROP_CLAIM` | A drop is claimed. |
| `DROP_STATUS` | Drop progress, with `show_drops_progress` on. Frequent. |
| `CHAT_MENTION` | Someone mentions you in a chat the miner is in. |
| `BET_START` | A prediction opened and a bet is scheduled. |
| `BET_FILTERS` | A bet was skipped by a filter or a limit. Frequent. |
| `BET_GENERAL` | A bet was skipped for another reason. |
| `BET_FAILED` | Twitch rejected a bet. |
| `BET_WIN`, `BET_LOSE`, `BET_REFUND` | The result of a bet. |

## What a message looks like

A message is a headline and one line of details, written from your side:

```
➕ +10 points on Alice
Watching Valorant · Balance 12.4k points
```

Points are always your points on that channel. On Telegram the headline is bold; on Discord it is the title of the embed. The glyph in front shows whether the news is good, bad or neutral, and is sent even with console `emojis` off.

## When a message does not arrive

A delivery problem never stops the miner. The first failure is printed to the console once, with the reason the service gave; later failures are not logged. The token and the webhook URL are never printed.

See [troubleshooting](troubleshooting.md#notifications-do-not-arrive) for the usual causes.
