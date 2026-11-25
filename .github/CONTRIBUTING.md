# Contributing

Bug reports and pull requests are welcome. For anything larger than a fix, open an [issue](https://github.com/kWAYTV/twitch-points-miner/issues) first so the approach can be agreed before you write it.

## Before you push

```sh
sh ci/check.sh                        # format, vet, tidy, build and test, as CI does
git config core.hooksPath ci/hooks    # once per clone: run it automatically
```

- `pre-commit` formats and vets the Go files being committed.
- `pre-push` runs `ci/check.sh`, the same checks CI runs.

CI runs lint, the tests (race detector, random order) and one build per platform side by side. Each job has a time budget. Do not retry a failing check, skip or loosen one, or raise a budget without a measured reason.

## Layout

Folders are named after what the code is about, and each file has one job.

| Folder | What it holds |
|---|---|
| `cmd/twitch-miner` | The entry point: paths, console setup and `config.json` loading. |
| `internal/miner` | The session, choosing who to watch, watch streaks, points, presence and identity repair. |
| `internal/streak` | Watch-streak rules that stand on their own, and the warm-start cache. |
| `internal/drops` | Claiming drops and reporting their progress. |
| `internal/twitch` | The Twitch client: GQL, users, streams, minute-watched, drops inventory, channel points. |
| `internal/twitchapi` | Endpoints, client identity, persisted GQL operations, user agents. |
| `internal/auth` | Device login and the cookie store. |
| `internal/pubsub` | The PubSub connection and what each message does. |
| `internal/prediction` | Prediction events: deciding, filtering and formatting bets. |
| `internal/chat` | The IRC chat client and when to be present in a chat. |
| `internal/domain` | Streamers, streams, bet and streamer settings, community goals. |
| `internal/events` | The event types the miner reports, with their polarity and glyph. |
| `internal/logging` | The console logger; notifiers and the dashboard subscribe to its events. |
| `internal/notify` | Discord and Telegram messages. |
| `internal/dashboard` | The analytics store, its HTTP API and the web page. |
| `internal/project` | The project's name, version and repository. |
| `internal/privacy`, `internal/update` | Log anonymisation; the self-updater. |
| `internal/ansi`, `internal/jsonfile`, `internal/jsonvalue`, `internal/random` | Console colours; writing JSON files; reading loose JSON values; nonces and jitter. |

## Conventions

- Keep files small and single-purpose, and functions short.
- Put new code in the package that owns its subject; add a package when a subject has none.
- Write a test with a change in behaviour. Code that talks to Twitch is tested against a local stand-in, never the real service.
- Document a new option in `docs/configuration.md`, and explain it on the topic page it belongs to.

## Licence

By contributing you agree that your work is released under the project's [CC BY-NC 4.0](../LICENSE) licence.
