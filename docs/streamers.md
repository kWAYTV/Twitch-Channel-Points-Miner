# Streamers and watching

## Who gets mined

```json
"streamers": ["streamer1", "streamer2"],
"streamers_exclude": ["streamer3"]
```

- `streamers` is a list of channel logins. Leave it empty to mine everyone you follow.
- `streamers_exclude` removes logins from that list. It is mostly useful with an empty `streamers`.

At startup the miner loads every streamer and reports once: `19 streamers loaded in 3.3 seconds: 3 online (…), 16 offline`. After that it reports each streamer going online or offline.

## The two watch slots

Twitch credits watch time on at most two streams at once, so the miner watches up to two live streamers at a time. Points for claiming bonus chests, raids and predictions are collected on every streamer in the list, watched or not.

`watch_priority` decides who gets the slots. The rules are applied in the order you list them:

```json
"watch_priority": ["STREAK", "DROPS", "ORDER"]
```

| Value | Prefers |
|---|---|
| `STREAK` | Streamers whose watch streak for the current stream has not been earned yet. |
| `DROPS` | One streamer whose stream carries drops. See [Drops](#drops). |
| `ORDER` | The order of your `streamers` list: the higher a streamer is, the sooner they get a slot. |
| `SUBSCRIBED` | Channels where you have a points multiplier. |
| `POINTS_ASCENDING` | The lowest balance first. |
| `POINTS_DESCENDING` | The highest balance first. |

An unknown value, or an empty list, falls back to the default above.

Whatever the rule, two streamers that tie are taken in the order of your `streamers` list.

Turn on `show_watch_queue` to have the miner explain which streamers got the slots and why.

### Games

```json
"game_priority": ["Counter-Strike", "VALORANT"],
"game_exclude": ["Just Chatting"]
```

- `game_priority` prefers streamers playing these games, the first game the most. It also decides which drops are worth a slot.
- `game_exclude` never watches a streamer while they play one of these games, and never claims their drops.

Names are matched without regard to upper or lower case, and have to be the game's name as Twitch shows it.

## Drops

Twitch counts drop progress on one channel at a time, so `DROPS` gives one slot to a stream that carries drops and leaves the other to earn points elsewhere.

- With `game_priority` empty, any stream with drops qualifies.
- With `game_priority` set, only streams of those games do, the first game first. A drop of a game you did not list never takes a slot.
- Finished drops are claimed every 30 minutes, whatever the game, except games in `game_exclude`.

A drop of a game your Twitch account is not connected to cannot be handed over, so the miner leaves that game alone: it does not claim its drops and does not give it the drops slot. It says so once, naming the game and where to connect the account, and picks the game up again by itself once the account is connected.

Twitch only reports the connection for campaigns you have some progress in, so the first minutes of a new campaign are watched before the miner can know.

## Watch streaks

Twitch awards a watch streak bonus for watching a channel's streams on consecutive occasions. With `STREAK` in `watch_priority`, a streamer who has just gone live and has not paid out the streak yet is moved to the front until the bonus arrives or enough minutes have been watched.

The miner saves what it knows about each streak to `log/watch_streak_cache.<username>.json`, so a restart does not make it chase streaks it already earned.

Streaks can be turned off for one streamer with `"watch_streak": false` in an [override](#per-streamer-overrides).

## What is claimed

| Key | Default | What it does |
|---|---|---|
| `claim_drops` | `true` | Claims finished drops from your inventory at startup and every 30 minutes. |
| `claim_moments` | `true` | Claims Twitch Moments when a streamer triggers one. |
| `follow_raid` | `true` | Joins a raid, which pays a bonus. |
| `community_goals` | `false` | Contributes points to a channel's active community goals, up to the per-stream limit Twitch sets and what the goal still needs. This spends your points. |

Bonus chests are always claimed.

`show_drops_progress` logs the progress of each drop at every inventory check.

## Chat presence

`chat_presence` sets when the miner joins a streamer's chat:

| Value | Joins |
|---|---|
| `ONLINE` | While the streamer is live (default). |
| `OFFLINE` | While the streamer is offline. |
| `ALWAYS` | All the time. |
| `NEVER` | Never. |

The miner only reads chat; it never writes. When someone mentions you, it reports a `CHAT_MENTION`. With `mention_without_at` on, your plain username counts as a mention, not only `@username`.

## Per-streamer overrides

`streamer_overrides` changes settings for one streamer. It is keyed by login, and anything left out is inherited from the global settings.

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

| Key | Overrides |
|---|---|
| `make_predictions` | `make_predictions` |
| `follow_raid` | `follow_raid` |
| `claim_drops` | `claim_drops` |
| `claim_moments` | `claim_moments` |
| `watch_streak` | Whether to chase this streamer's watch streak. |
| `community_goals` | `community_goals` |
| `chat_presence` | `chat_presence` |
| `bet` | Any of the [prediction settings](predictions.md). |

The file is written with the streamers in alphabetical order.
