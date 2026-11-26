# Dashboard

A web page that shows, for each streamer, your balance, a chart of points over time with the events behind it, and an activity log. It refreshes itself every 45 seconds and works on a phone.

```json
"dashboard": {
  "enabled": true,
  "host": "127.0.0.1",
  "port": 5000
}
```

It is on by default at <http://localhost:5000>. Set `enabled` to `false` to turn it off; history is then no longer recorded either.

## Reaching it from another machine

With the default `host` of `127.0.0.1` the page only answers on the machine the miner runs on. Set `host` to `0.0.0.0` to answer on every network interface.

The environment variable `TCPM_DASHBOARD_ADDR` (for example `0.0.0.0:8080`) overrides both `host` and `port`. Its older name, `TCPM_ANALYTICS_ADDR`, still works; when both are set, `TCPM_DASHBOARD_ADDR` is used. The Docker image sets the address to `0.0.0.0:5000` so the page is reachable through the published port, whatever the config says.

**The dashboard has no login.** Anyone who can reach the port can read your balances and activity. Keep it on your own network or behind a reverse proxy that adds authentication, and never open the port to the internet. In `compose.yaml`, publishing the port as `"127.0.0.1:5000:5000"` keeps it on the Docker host.

## History

History is stored in the `analytics/` folder as plain JSONL files, per streamer. A balance snapshot is added every 5 minutes, and each gain, bet and claim as it happens.

Every bet is also kept in `analytics/bets.jsonl` once its result is known: the strategy, the stake, how the voters and the pool stood, and what it won or lost. Under the activity log, **Bets by strategy** adds them up, so you can see whether a strategy is actually earning points.

- To reset one streamer, delete their files in `analytics/`.
- To start over, delete the folder.
- To keep history across reinstalls, keep the folder (in Docker it is inside `./data`).

The dashboard shows real streamer names even with `anonymize_logs` on; that option only affects the console and the log file.

## JSON API

The page reads three endpoints, which you can use yourself:

| Request | Returns |
|---|---|
| `GET /api/streamers` | Every streamer with their current balance. |
| `GET /api/bets` | How each bet strategy has done over every settled bet, and the 50 newest bets with the figures they were decided from. |
| `GET /api/streamer/{name}?range=24h` | One streamer's balance series and events. `range` is `24h` (default), `7d`, `30d` or `all`. |
