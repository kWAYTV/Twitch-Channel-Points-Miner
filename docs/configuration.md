# Configuration

`config.json` is created on the first run and filled in with defaults; keys added by newer versions are appended to an existing file. Only `username` has to be set.

The miner uses `./config.json` if it exists in the current directory, otherwise the directory of the executable. `cookies/`, `log/` and `analytics/` live next to it. Override the locations with `-config`, `-data-dir`, `TCPM_CONFIG` or `TCPM_DATA_DIR`.

## Account

| Key | Default | Meaning |
|---|---|---|
| `username` | | Twitch login to mine with. |
| `password` | `""` | Not needed; login uses Twitch's device flow. |
| `auto_update` | `true` | Replace the binary when a newer release is published. Always off in the Docker image. |

## Streamers

| Key | Default | Meaning |
|---|---|---|
| `streamers` | `[]` | Channel logins to mine. Empty means everyone you follow. |
| `streamers_exclude` | `[]` | Logins to skip, useful together with an empty `streamers`. |
| `watch_priority` | `["STREAK", "DROPS", "ORDER"]` | How the two watch slots are filled. Also `SUBSCRIBED`, `POINTS_ASCENDING`, `POINTS_DESCENDING`. |
| `game_priority` | `[]` | Games to prefer, in order. |
| `game_exclude` | `[]` | Games never to watch. |
| `streamer_overrides` | `{}` | Per-streamer settings, see below. |

## What to claim

| Key | Default | Meaning |
|---|---|---|
| `claim_drops` | `true` | Claim drops from the inventory every 30 minutes. |
| `claim_drops_startup` | `true` | Also claim once at startup. |
| `claim_moments` | `true` | Claim Twitch Moments. |
| `follow_raid` | `true` | Join raids and keep mining the destination. |
| `community_goals` | `false` | Contribute points to community goals. |
| `betting(make_predictions)` | `true` | Place predictions. |
| `chat_presence` | `"ONLINE"` | When to sit in a streamer's chat: `ALWAYS`, `NEVER`, `ONLINE` or `OFFLINE`. |
| `disable_at_in_nickname` | `false` | Treat a plain username in chat as a mention, without `@`. |
| `watch_streak_warm_start_cache` | `true` | Remember watch-streak state across restarts. |

## Predictions

Settings under `bet`:

| Key | Default | Meaning |
|---|---|---|
| `strategy` | `"SMART"` | How the outcome is chosen. Others include `MOST_VOTED`, `HIGH_ODDS`, `PERCENTAGE` and `SMART_MONEY`. |
| `percentage` | `5` | Share of the balance to bet. |
| `percentage_gap` | `20` | Minimum gap between outcomes before betting. |
| `max_points` | `50000` | Most points placed on one bet. |
| `minimum_points` | `0` | Do not bet while the balance is below this. |
| `stealth_mode` | `false` | Keep the stake just under the highest bet. |
| `deduct_stake_on_place` | `true` | Subtract the stake from the shown balance when the bet is placed. |
| `delay_mode`, `delay` | `"FROM_END"`, `6` | When to place the bet: seconds from the start or the end of the window. |
| `filter_condition` | none | Skip a bet unless a live figure meets a rule, for example `{ "by": "TOTAL_USERS", "where": "GTE", "value": 50 }`. `by` is one of `PERCENTAGE_USERS`, `ODDS`, `ODDS_PERCENTAGE`, `TOP_POINTS`, `TOTAL_USERS`, `TOTAL_POINTS`; `where` is `GT`, `LT`, `GTE` or `LTE`. |

## Per-streamer overrides

`streamer_overrides` is keyed by login. Anything left out is inherited from the global settings.

```json
"streamer_overrides": {
  "streamer1": {
    "make_predictions": false
  },
  "streamer2": {
    "claim_drops": false,
    "watch_streak": true,
    "bet": { "max_points": 5000, "strategy": "PERCENTAGE" }
  }
}
```

Keys: `make_predictions`, `follow_raid`, `claim_drops`, `claim_moments`, `watch_streak`, `community_goals`, `chat_presence` and `bet`.

## Dashboard

```json
"analytics": { "enabled": true, "host": "127.0.0.1", "port": 5000 }
```

- Set `host` to `0.0.0.0` to reach it from other machines. `TCPM_ANALYTICS_ADDR` (for example `0.0.0.0:8080`) overrides both `host` and `port`.
- History is stored as JSONL files in `analytics/`; delete a streamer's files to reset their history.
- The dashboard has no login and shows real streamer names even with `privacy.anonymize_logs`.
- JSON API: `GET /api/streamers` and `GET /api/streamer/{name}?range=24h|7d|30d|all`.

## Notifications

```json
"discord": { "webhook_api": "https://discord.com/api/webhooks/...", "events": [] },
"telegram": { "chat_id": 123456789, "token": "123456:ABC...", "events": ["STREAMER_ONLINE", "BET_WIN"], "disable_notification": false }
```

- **Discord:** create a webhook in the channel settings and paste its URL.
- **Telegram:** create a bot with [@BotFather](https://t.me/BotFather), send it a message once, and set `chat_id` to your numeric chat id. `@channelname` works for public channels.
- An empty `webhook_api`, or an empty `token` or `chat_id`, turns that notifier off.
- An empty `events` list sends everything except three high-volume events: `GAIN_FOR_WATCH`, `DROP_STATUS` and `BET_FILTERS`. `["*"]` sends those too.
- The first delivery failure is printed once; tokens and webhook URLs are never logged.

Events: `STARTUP`, `SHUTDOWN`, `STREAMER_ONLINE`, `STREAMER_OFFLINE`, `GAIN_FOR_RAID`, `GAIN_FOR_CLAIM`, `GAIN_FOR_WATCH`, `GAIN_FOR_WATCH_STREAK`, `BET_WIN`, `BET_LOSE`, `BET_REFUND`, `BET_FILTERS`, `BET_GENERAL`, `BET_FAILED`, `BET_START`, `BONUS_CLAIM`, `MOMENT_CLAIM`, `JOIN_RAID`, `DROP_CLAIM`, `DROP_STATUS`, `CHAT_MENTION`.

## Console and logs

| Key | Default | Meaning |
|---|---|---|
| `emojis` | `true` | Prefix console lines with a glyph. Notifications always carry one. |
| `show_seconds` | `false` | Include seconds in timestamps. |
| `show_game` | `true` | Show what a streamer is playing. |
| `show_drops_progress` | `false` | Log drop progress at each inventory check. |
| `watch_queue_logging` | `false` | Explain which streamers got the watch slots and why. |
| `save_logs` | `false` | Also write the console output to `log/<username>.log`. |
| `timezone` | auto | IANA name such as `Europe/Madrid`, for systems that cannot detect it. |
| `privacy.anonymize_logs` | `false` | Replace streamer names and balances in logs with stand-ins. |
| `debug`, `debug_deep` | `false` | Request summaries, and full payloads. Deep debug is ignored while logs are anonymised. |
| `disable_ssl_cert_verification` | `false` | Skip TLS verification for the PubSub connection only. Leave off. |

`smart_logging`, `show_username_in_console` and `show_claimed_bonus_msg` are still accepted in the file but currently have no effect.

## Environment variables

| Variable | Meaning |
|---|---|
| `TCPM_CONFIG` | Path of `config.json`. |
| `TCPM_DATA_DIR` | Folder for `config.json`, `cookies/`, `log/` and `analytics/`. |
| `TCPM_ANALYTICS_ADDR` | Dashboard listen address, overriding the config. |
| `TCPM_DISABLE_AUTO_UPDATE` | `1` turns the self-updater off whatever the config says. |
| `TZ` | Time zone for timestamps in a container. |
