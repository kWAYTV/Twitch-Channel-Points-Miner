# Predictions

With `"make_predictions": true` the miner bets your channel points on a streamer's predictions. Bets can lose: set limits you are comfortable with, or turn it off.

All settings below live under `bet`. They can be changed for one streamer under `streamer_overrides.<login>.bet` (see [per-streamer overrides](streamers.md#per-streamer-overrides)).

```json
"make_predictions": true,
"bet": {
  "strategy": "SMART",
  "percentage": 5,
  "percentage_gap": 20,
  "max_points": 50000,
  "minimum_points": 0,
  "stealth_mode": false,
  "delay_mode": "FROM_END",
  "delay": 6,
  "filter_condition": { "by": null, "where": null, "value": null }
}
```

## Choosing the outcome

`strategy` picks which outcome to bet on, using the figures at the moment the bet is placed.

| Strategy | Picks |
|---|---|
| `SMART` | The outcome most people voted for. If the two most-voted outcomes are closer than `percentage_gap` percentage points, the one with the highest odds instead. |
| `MOST_VOTED` | The outcome with the most voters. |
| `HIGH_ODDS` | The outcome with the highest odds (the biggest payout, the least likely). |
| `PERCENTAGE` | The outcome with the highest odds percentage. |
| `SMART_MONEY` | The outcome holding the largest single bet. |
| `NUMBER_1` … `NUMBER_8` | Always the outcome in that position. If the prediction has fewer outcomes, the one with the highest odds. |

`percentage_gap` (default `20`) is only used by `SMART`.

## Sizing the stake

1. Start from `percentage` percent of your balance on that channel (default `5`).
2. Cap it at `max_points` (default `50000`).
3. With `stealth_mode` on, if the stake would match or beat the highest bet on that outcome, lower it to one point under that bet, so yours is never the top bet.
4. Twitch does not accept less than 10 points. A smaller stake is raised to 10 when the balance allows it; otherwise the bet is skipped.

`minimum_points` skips betting while your balance on the channel is at or below it. Use it to keep a reserve.

The stake is subtracted from the balance the miner shows as soon as the bet is placed, without waiting for Twitch to report it.

## When the bet is placed

Later is better informed: the figures are read when the bet is placed, so waiting until the window is about to close sees almost all the votes.

| `delay_mode` | `delay` means |
|---|---|
| `FROM_END` | Seconds before the window closes (default, with `delay` `6`). |
| `FROM_START` | Seconds after the prediction opens. |
| `PERCENTAGE` | A fraction of the window, so `0.5` is halfway through. |

## Filters

`filter_condition` skips a bet unless a figure meets a rule. For example, only bet when at least 100 people have voted:

```json
"filter_condition": { "by": "TOTAL_USERS", "where": "GTE", "value": 100 }
```

| `by` | Figure compared |
|---|---|
| `TOTAL_USERS` | Voters across all outcomes. |
| `TOTAL_POINTS` | Points across all outcomes. |
| `DECISION_USERS` | Voters on the outcome that would be bet on. |
| `DECISION_POINTS` | Points on that outcome. |
| `PERCENTAGE_USERS` | Share of voters on that outcome. |
| `ODDS` | Odds of that outcome. |
| `ODDS_PERCENTAGE` | Odds percentage of that outcome. |
| `TOP_POINTS` | The largest single bet on that outcome. |

`where` is `GT` (greater than), `GTE` (greater or equal), `LT` (less than) or `LTE` (less or equal). Leaving the three fields `null` turns the filter off.

## What you see

| Event | Meaning |
|---|---|
| `BET_START` | A prediction opened and the bet is scheduled. |
| `BET_FILTERS` | The bet was skipped by a filter, by `minimum_points`, or because the stake was under Twitch's minimum. |
| `BET_GENERAL` | The bet was skipped for another reason, such as no outcome matching the strategy. |
| `BET_FAILED` | Twitch rejected the bet. |
| `BET_WIN`, `BET_LOSE`, `BET_REFUND` | The result. |

These can be sent to Discord or Telegram: see [notifications](notifications.md).
