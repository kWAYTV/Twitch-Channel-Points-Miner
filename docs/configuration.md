# Configuration

Every key in `config.json`, in the order the file is written. Only `username` has to be set. The topic pages explain how the settings behave; this page is the reference.

## The file

- It is created on the first run with defaults. [config.example.json](../config.example.json) is that same file.
- It is written in the groups below, with a blank line between them.
- A key added by a newer version is filled in with its default, in its place.
- A file in another layout is rewritten once in this one. Every value is kept.
- A file from an older version is converted when it is loaded, and rewritten once. See [settings from older versions](#settings-from-older-versions).
- A key the miner does not know is kept, at the end of the file, and ignored.

Where the file is looked for is described in [getting started](getting-started.md#where-files-are-kept).

## Account

| Key | Default | Meaning |
|---|---|---|
| `username` | | Twitch login to mine with. |
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
| `game_exclude` | `[]` | Games never to watch, and whose drops are never claimed. |

## What to claim

Explained in [streamers and watching](streamers.md#what-is-claimed).

| Key | Default | Meaning |
|---|---|---|
| `claim_drops` | `true` | Claim drops from the inventory at startup and every 30 minutes. |
| `claim_moments` | `true` | Claim Twitch Moments. |
| `follow_raid` | `true` | Join raids. |
| `community_goals` | `false` | Contribute points to community goals. |
| `chat_presence` | `"ONLINE"` | When to sit in a streamer's chat: `ALWAYS`, `NEVER`, `ONLINE` or `OFFLINE`. |
| `mention_without_at` | `false` | Treat a plain username in chat as a mention, without `@`. |

## Predictions

Explained in [predictions](predictions.md).

| Key | Default | Meaning |
|---|---|---|
| `make_predictions` | `true` | Place predictions. |
| `bet.strategy` | `"SMART"` | How the outcome is chosen: `SMART`, `MOST_VOTED`, `HIGH_ODDS`, `PERCENTAGE`, `SMART_MONEY`, `NUMBER_1` to `NUMBER_8`. |
| `bet.percentage` | `5` | Percent of the balance to bet. |
| `bet.percentage_gap` | `20` | For `SMART`: how close the two most-voted outcomes must be to switch to the highest odds. |
| `bet.max_points` | `50000` | Most points placed on one bet. |
| `bet.minimum_points` | `0` | Do not bet while the balance is at or below this. |
| `bet.stealth_mode` | `false` | Keep the stake under the highest bet on the outcome. |
| `bet.delay_mode` | `"PERCENTAGE"` | What `delay` counts from: `PERCENTAGE`, `FROM_START` or `FROM_END`. |
| `bet.delay` | `0.5` | A fraction of the window with `PERCENTAGE` (`0.5` is halfway), otherwise seconds. |
| `bet.filter_condition` | off | Skip a bet unless a figure meets a rule: `by`, `where`, `value`. |

## Notifications

Explained in [notifications](notifications.md).

| Key | Default | Meaning |
|---|---|---|
| `discord.webhook` | `""` | The webhook URL. Empty turns Discord off. |
| `discord.events` | `[]` | Events to send. Empty sends all but the frequent ones. |
| `telegram.token` | `""` | The bot token from @BotFather. |
| `telegram.chat_id` | `""` | Your numeric chat id, or `@channelname`. |
| `telegram.events` | `[]` | Events to send. Empty sends all but the frequent ones. |
| `telegram.silent` | `false` | Deliver silently. |
| `matrix.homeserver` | `""` | Address of the homeserver, such as `https://matrix.org`. |
| `matrix.access_token` | `""` | Access token of the account that sends the messages. |
| `matrix.room_id` | `""` | The room's internal id, which starts with `!`. Matrix is off unless all three are set. |
| `matrix.events` | `[]` | Events to send. Empty sends all but the frequent ones. |

## Dashboard

Explained in [dashboard](dashboard.md).

| Key | Default | Meaning |
|---|---|---|
| `dashboard.enabled` | `true` | Record history and serve the web page. |
| `dashboard.host` | `"127.0.0.1"` | Address to listen on. `0.0.0.0` answers on every interface. |
| `dashboard.port` | `5000` | Port to listen on. |

## Console and logs

| Key | Default | Meaning |
|---|---|---|
| `emojis` | `true` | Prefix console lines with a glyph. Notifications always carry one. |
| `show_seconds` | `false` | Include seconds in timestamps. |
| `show_username` | `false` | Put the account's login in every console and log line, after the timestamp: `[INFO] 10:23 24/10/26: [my_username] ...`. Left out while `anonymize_logs` is on, and never part of a notification. |
| `show_game` | `true` | Show what a streamer is playing. |
| `show_drops_progress` | `false` | Log drop progress at each inventory check. |
| `show_watch_queue` | `false` | Explain which streamers got the watch slots and why. |
| `save_logs` | `false` | Also write the console output to `log/<username>.log`. |
| `timezone` | `null` (detect) | IANA name such as `Europe/Madrid`, for systems that cannot detect it. |
| `anonymize_logs` | `false` | Replace streamer names and balances in logs with stand-ins. Also leaves out the login `show_username` adds. |
| `debug` | `false` | `true` logs a summary of each request. `"full"` also logs full payloads, which are left out while logs are anonymised. |

## Settings from older versions

A file written for an older version keeps working. It is converted when it is loaded and rewritten once with the names below. Nothing has to be edited by hand.

Renamed:

| Old key | New key |
|---|---|
| `betting(make_predictions)` | `make_predictions` |
| `analytics` | `dashboard` |
| `discord.webhook_api` | `discord.webhook` |
| `telegram.disable_notification` | `telegram.silent` |
| `privacy.anonymize_logs` | `anonymize_logs` |
| `disable_at_in_nickname` | `mention_without_at` |
| `watch_queue_logging` | `show_watch_queue` |
| `show_username_in_console` | `show_username`. It did nothing in older versions; it now does what it says. |
| `bet.delay_mode` `FROM_END` with `bet.delay` `6` | `PERCENTAGE` with `0.5`, the new default: see [when the bet is placed](predictions.md#when-the-bet-is-placed). |
| `debug` with `debug_deep` | `debug`: `false`, `true` or `"full"`. `"full"` is what `debug` and `debug_deep` both on used to be. |

If a file has both the old and the new key, the new one is used.

Removed, because the behaviour is now fixed:

| Old key | Now |
|---|---|
| `password` | Was not used. [Login](getting-started.md#logging-in) is by code. |
| `claim_drops_startup` | Drops are claimed at startup whenever `claim_drops` is on. |
| `watch_streak_warm_start_cache` | Watch-streak state is always remembered across restarts. |
| `bet.deduct_stake_on_place` | The stake always leaves the shown balance when the bet is placed. |
| `disable_ssl_cert_verification` | Certificates are always verified. |

`smart_logging` and `show_claimed_bonus_msg` did nothing and were removed earlier. They are not converted: a file that still has them keeps them at the end, and they are ignored.

## Environment variables

These override the file and are mostly useful in containers.

| Variable | Meaning |
|---|---|
| `TCPM_CONFIG` | Path of `config.json`. |
| `TCPM_DATA_DIR` | Folder for `config.json`, `cookies/`, `log/` and `analytics/`. |
| `TCPM_ANALYTICS_ADDR` | Dashboard listen address, for example `0.0.0.0:8080`. Overrides `dashboard.host` and `dashboard.port`. |
| `TCPM_DISABLE_AUTO_UPDATE` | `1` turns the self-updater off whatever the config says. |
| `TZ` | Time zone for timestamps in a container. |

The command-line flags `-config <file>` and `-data-dir <folder>` take priority over the variables.
