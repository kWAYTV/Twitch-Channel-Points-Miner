# Configuration

Every key in `config.json`, in the order the file is written. Only `username` has to be set. The topic pages explain how the settings behave; this page is the reference.

## The file

- It is created on the first run with defaults. [config.example.json](../config.example.json) is that same file.
- It is written in the groups below, with a blank line between them.
- A key added by a newer version is filled in with its default, in its place.
- A file in another layout is rewritten once in this one. Every value is kept.
- A key the miner does not know is kept, at the end of the file, and ignored.

Where the file is looked for is described in [getting started](getting-started.md#where-files-are-kept).

## Account

| Key | Default | Meaning |
|---|---|---|
| `username` | | Twitch login to mine with. |
| `password` | `""` | Not needed: [login](getting-started.md#logging-in) uses a code, not your password. |
| `auto_update` | `true` | Replace the binary when a newer release is published. Always off in the Docker image. |

## Streamers

Explained in [streamers and watching](streamers.md).

| Key | Default | Meaning |
|---|---|---|
| `streamers` | `[]` | Channel logins to mine. Empty means everyone you follow. |
| `streamers_exclude` | `[]` | Logins to skip. |
| `streamer_overrides` | `{}` | Settings for one streamer, keyed by login. |
| `watch_priority` | `["STREAK", "DROPS", "ORDER"]` | How the two watch slots are filled. Also `SUBSCRIBED`, `POINTS_ASCENDING`, `POINTS_DESCENDING`. |
| `game_priority` | `[]` | Games to prefer, in order. |
| `game_exclude` | `[]` | Games never to watch. |

## What to claim

Explained in [streamers and watching](streamers.md#what-is-claimed).

| Key | Default | Meaning |
|---|---|---|
| `claim_drops` | `true` | Claim drops from the inventory every 30 minutes. |
| `claim_drops_startup` | `true` | Also claim once at startup. |
| `claim_moments` | `true` | Claim Twitch Moments. |
| `follow_raid` | `true` | Join raids. |
| `community_goals` | `false` | Contribute points to community goals. |
| `chat_presence` | `"ONLINE"` | When to sit in a streamer's chat: `ALWAYS`, `NEVER`, `ONLINE` or `OFFLINE`. |
| `disable_at_in_nickname` | `false` | Treat a plain username in chat as a mention, without `@`. |
| `watch_streak_warm_start_cache` | `true` | Remember watch-streak state across restarts. |

## Predictions

Explained in [predictions](predictions.md).

| Key | Default | Meaning |
|---|---|---|
| `betting(make_predictions)` | `true` | Place predictions. |
| `bet.strategy` | `"SMART"` | How the outcome is chosen: `SMART`, `MOST_VOTED`, `HIGH_ODDS`, `PERCENTAGE`, `SMART_MONEY`, `NUMBER_1` to `NUMBER_8`. |
| `bet.percentage` | `5` | Percent of the balance to bet. |
| `bet.percentage_gap` | `20` | For `SMART`: how close the two most-voted outcomes must be to switch to the highest odds. |
| `bet.max_points` | `50000` | Most points placed on one bet. |
| `bet.minimum_points` | `0` | Do not bet while the balance is at or below this. |
| `bet.stealth_mode` | `false` | Keep the stake under the highest bet on the outcome. |
| `bet.deduct_stake_on_place` | `true` | Subtract the stake from the shown balance when the bet is placed. |
| `bet.delay_mode` | `"FROM_END"` | What `delay` counts from: `FROM_END`, `FROM_START` or `PERCENTAGE`. |
| `bet.delay` | `6` | Seconds, or a fraction of the window with `PERCENTAGE`. |
| `bet.filter_condition` | off | Skip a bet unless a figure meets a rule: `by`, `where`, `value`. |

## Notifications

Explained in [notifications](notifications.md).

| Key | Default | Meaning |
|---|---|---|
| `discord.webhook_api` | `""` | The webhook URL. Empty turns Discord off. |
| `discord.events` | `[]` | Events to send. Empty sends all but the frequent ones. |
| `telegram.token` | `""` | The bot token from @BotFather. |
| `telegram.chat_id` | `""` | Your numeric chat id, or `@channelname`. |
| `telegram.events` | `[]` | Events to send. Empty sends all but the frequent ones. |
| `telegram.disable_notification` | `false` | Deliver silently. |

## Dashboard

Explained in [dashboard](dashboard.md).

| Key | Default | Meaning |
|---|---|---|
| `analytics.enabled` | `true` | Record history and serve the web page. |
| `analytics.host` | `"127.0.0.1"` | Address to listen on. `0.0.0.0` answers on every interface. |
| `analytics.port` | `5000` | Port to listen on. |

## Console and logs

| Key | Default | Meaning |
|---|---|---|
| `emojis` | `true` | Prefix console lines with a glyph. Notifications always carry one. |
| `show_seconds` | `false` | Include seconds in timestamps. |
| `show_game` | `true` | Show what a streamer is playing. |
| `show_drops_progress` | `false` | Log drop progress at each inventory check. |
| `watch_queue_logging` | `false` | Explain which streamers got the watch slots and why. |
| `save_logs` | `false` | Also write the console output to `log/<username>.log`. |
| `timezone` | `null` (detect) | IANA name such as `Europe/Madrid`, for systems that cannot detect it. |
| `privacy.anonymize_logs` | `false` | Replace streamer names and balances in logs with stand-ins. |
| `debug` | `false` | Log a summary of each request. |
| `debug_deep` | `false` | Log full payloads. Ignored while logs are anonymised. |
| `disable_ssl_cert_verification` | `false` | Skip TLS verification for the PubSub connection only. Leave off. |

`smart_logging`, `show_username_in_console` and `show_claimed_bonus_msg` did nothing and were removed. An old file may still contain them; they are ignored.

## Environment variables

These override the file and are mostly useful in containers.

| Variable | Meaning |
|---|---|
| `TCPM_CONFIG` | Path of `config.json`. |
| `TCPM_DATA_DIR` | Folder for `config.json`, `cookies/`, `log/` and `analytics/`. |
| `TCPM_ANALYTICS_ADDR` | Dashboard listen address, for example `0.0.0.0:8080`. Overrides `analytics.host` and `analytics.port`. |
| `TCPM_DISABLE_AUTO_UPDATE` | `1` turns the self-updater off whatever the config says. |
| `TZ` | Time zone for timestamps in a container. |

The command-line flags `-config <file>` and `-data-dir <folder>` take priority over the variables.
