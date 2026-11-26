# Troubleshooting

## The login code expired, or I want another account

Restart the miner to get a new code. To log in from scratch or with a different account, delete `cookies/<username>.json` and restart. In Docker the file is in `./data/cookies/`.

## The container cannot write to `./data`

The container runs as user 1000. If the folder belongs to another user, either hand it over:

```sh
sudo chown -R 1000:1000 ./data
```

or run the container as your own user:

```sh
PUID=$(id -u) PGID=$(id -g) docker compose up -d
```

## The miner stops with `invalid config`

`config.json` is not valid JSON; the message says where. The usual causes are a missing or trailing comma and a missing quote. If you would rather start again, move the file away and the next start writes a fresh one.

## I cannot open the dashboard

- **On the same machine:** check `dashboard.enabled` is `true` and nothing else uses port 5000. Change `dashboard.port` if something does.
- **From another machine:** outside Docker the page only listens on `127.0.0.1`. Set `dashboard.host` to `0.0.0.0`.
- **In Docker:** check the port is published (`docker compose ps`), and that it is not published on `127.0.0.1` only.

More in [dashboard](dashboard.md).

## Notifications do not arrive

The first delivery failure is printed to the console once, with the reason.

- **Telegram, "chat not found":** send your bot a message first, and check `chat_id` is your numeric id.
- **Telegram, "Unauthorized":** the token is wrong or was revoked. Get a new one from @BotFather.
- **Discord:** the webhook was deleted, or the URL is incomplete.
- **Some events are missing:** an empty `events` list leaves out `GAIN_FOR_WATCH`, `DROP_STATUS` and `BET_FILTERS`. Use `["*"]` to get everything.

More in [notifications](notifications.md).

## Timestamps are in the wrong time zone

Set `timezone` in `config.json` to an IANA name such as `Europe/Madrid`. In Docker, set `TZ` instead.

## I need more detail in the logs

- `show_watch_queue` explains who is being watched and why.
- `save_logs` also writes the console output to `log/<username>.log`.
- `"debug": true` adds a summary of each request. `"debug": "full"` also adds full payloads, which include private data: do not share them unredacted. Full payloads are left out while `anonymize_logs` is on.

When reporting a bug, turn on `anonymize_logs` to replace streamer names and balances with stand-ins before you paste a log.
