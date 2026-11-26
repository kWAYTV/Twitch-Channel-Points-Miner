# Predictions

With `"make_predictions": true` the miner bets your channel points on a streamer's predictions. Bets can lose: set limits you are comfortable with, or turn it off.

All settings below live under `bet`. They can be changed for one streamer under `streamer_overrides.<login>.bet` (see [per-streamer overrides](streamers.md#per-streamer-overrides)).

```json
"make_predictions": true,
"bet": {
  "strategy": "KELLY",
  "percentage": 5,
  "percentage_gap": 20,
  "max_points": 50000,
  "minimum_points": 0,
  "stealth_mode": false,
  "delay_mode": "PERCENTAGE",
  "delay": 0.75,
  "filter_condition": { "by": null, "where": null, "value": null }
}
```

## Choosing the outcome

`strategy` picks which outcome to bet on, using the figures at the moment the bet is placed.

| Strategy | Picks |
|---|---|
| `KELLY` | The outcome the pool underprices, and only when there is one. Default. See [the Kelly strategy](#the-kelly-strategy). |
| `SMART` | The outcome most people voted for. If the two most-voted outcomes are closer than `percentage_gap` percentage points, the one with the highest odds instead. |
| `MOST_VOTED` | The outcome with the most voters. |
| `HIGH_ODDS` | The outcome with the highest odds (the biggest payout, the least likely). |
| `PERCENTAGE` | The outcome with the highest odds percentage. |
| `SMART_MONEY` | The outcome holding the largest single bet. |
| `NUMBER_1` … `NUMBER_8` | Always the outcome in that position. If the prediction has fewer outcomes, the one with the highest odds. |

`percentage_gap` (default `20`) is only used by `SMART`.

## The Kelly strategy

A Twitch prediction is a parimutuel pool: the winners share everything that was bet, in proportion to their stakes. A bet only pays off over time when the outcome is likelier than its share of the pool says. `KELLY` applies the two rules that follow from that:

- **Bet only on value.** The chance of an outcome is estimated from its share of the viewers who voted, one viewer one vote. The pool gives the odds. An outcome is underpriced when that chance is larger than its share of the pool. If none is underpriced enough, no bet is placed.
- **Size by the Kelly criterion.** The stake is the one that makes the balance grow fastest over many bets, worked out for a pool where your own stake lowers the payout. A quarter of that is bet, the usual precaution when the chance is an estimate, and it never exceeds `percentage` of the balance or `max_points`.

### Small and large predictions

A prediction with 6 voters and one with 600 are not the same bet: one voter moves the first by a sixth. `KELLY` handles this without a rule per size. Its estimate starts from the pool's own split and moves towards the voters' share as more of them vote, the pool counting for 20 voters:

| Voters, two thirds on Yes, pool split evenly | Estimated chance of Yes | Full Kelly stake |
|---|---|---|
| 6 | 54% | 8% of the balance |
| 60 | 63% | 25% |
| 600 | 66% | 32% |

So a small prediction gets a small stake or none, and a large one is judged by its voters. A quarter of the full stake is what is bet, within your limits. When even the full stake would be under Twitch's minimum of 10 points, no bet is placed.

The estimate is the weak part: nothing published shows that viewer counts predict Twitch outcomes better than the pool does. The strategy is disciplined about when and how much to bet; it does not make the estimate right.

`percentage_gap` is not used by `KELLY`.

## Sizing the stake

With `KELLY`, `percentage` and `max_points` are upper limits and the stake is usually smaller. With every other strategy:

1. Start from `percentage` percent of your balance on that channel (default `5`).
2. Cap it at `max_points` (default `50000`).
3. With `stealth_mode` on, if the stake would match or beat the highest bet on that outcome, lower it to one point under that bet, so yours is never the top bet.
4. Twitch does not accept less than 10 points. A smaller stake is raised to 10 when the balance allows it; otherwise the bet is skipped.

`minimum_points` skips betting while your balance on the channel is at or below it. Use it to keep a reserve.

Whatever the strategy, no bet is placed when every point in the pool is on the chosen outcome: a win would pay back the stake and nothing more.

The stake is subtracted from the balance the miner shows as soon as the bet is placed, without waiting for Twitch to report it.

## When the bet is placed

A prediction that is already open when the miner starts is picked up from its next update and bet on in the time that is left. One you already bet on yourself, from the site or another session, is left alone.

By default the bet is placed three quarters of the way through the window, give or take a tenth of the window with [`humanize`](configuration.md#humanize) on. Later is better informed, because the figures are read when the bet is placed, but a channel can close betting before the window ends, and a bet left for the last seconds is then never placed (`Skip bet ...: event status is LOCKED`).

| `delay_mode` | `delay` means |
|---|---|
| `PERCENTAGE` | A fraction of the window: `0.75`, the default, is three quarters through, and `0.5` halfway. |
| `FROM_START` | Seconds after the prediction opens. |
| `FROM_END` | Seconds before the window closes. `6` sees almost all the votes, and misses the bet when betting is closed early. |

Older versions bet six seconds before the end. A config from one of them is moved to the new default when it is converted; write `FROM_END` and `6` back if you prefer it.

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

These can be sent to Discord, Telegram or Matrix: see [notifications](notifications.md).
