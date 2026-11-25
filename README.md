![Twitch Channel Points Miner](docs/assets/banner.png)

# Twitch Points Miner

[![ci](https://github.com/kWAYTV/twitch-points-miner/actions/workflows/ci.yml/badge.svg)](https://github.com/kWAYTV/twitch-points-miner/actions/workflows/ci.yml)
[![docker](https://github.com/kWAYTV/twitch-points-miner/actions/workflows/docker.yml/badge.svg)](https://github.com/kWAYTV/twitch-points-miner/actions/workflows/docker.yml)
[![licence](https://img.shields.io/badge/licence-CC%20BY--NC%204.0-lightgrey)](LICENSE)

Earns Twitch channel points for you by watching streams in the background. Written in Go: one small binary, quick to start, light on CPU and memory.

## What it does

- Watches up to two live streams at a time and keeps watch streaks going.
- Claims bonus chests, drops and moments, and follows raids.
- Places predictions with configurable strategies and filters.
- Mines a list of streamers, or everyone you follow.
- Shows balances and history on a built-in web dashboard.
- Sends Discord and Telegram notifications in plain sentences.

## Quick start

### Docker

```sh
docker run -d --name twitch-miner --restart unless-stopped \
  -v twitch-miner-data:/data -p 5000:5000 \
  ghcr.io/kwaytv/twitch-points-miner:latest
```

The first start writes `/data/config.json`. Set `username` in it and restart the container. Then run `docker logs -f twitch-miner`, open <https://www.twitch.tv/activate> and enter the code from the logs. The login is saved, so this is needed once.

The image supports `linux/amd64` and `linux/arm64`, runs as UID 1000 and keeps everything in `/data`. A bind-mounted folder must be writable by that user (`chown -R 1000:1000 ./data`). Update with `docker pull`.

### From source

Requires Go 1.23 or newer.

```sh
go run ./cmd/twitch-miner
```

The first run creates `config.json`. Set `username`, run it again and complete the same device login. Press Ctrl+C to stop; a session summary is printed.

## Configuration

`config.json` is created with working defaults, so `username` is the only required change. Leave `streamers` empty to mine everyone you follow.

```json
{
  "username": "your-twitch-username",
  "streamers": ["streamer1", "streamer2"],
  "claim_drops": true,
  "betting(make_predictions)": true
}
```

Every option is described in [docs/configuration.md](docs/configuration.md).

## Dashboard

Open <http://localhost:5000> for each streamer's balance, a points-over-time chart with the events behind it, and an activity log. It is on by default and listens on `127.0.0.1` (the Docker image listens on all interfaces). It has no login, so do not expose the port to the internet.

## Notifications

Add a Discord webhook or a Telegram bot to `config.json` and choose which events to receive. Each message is a bold headline with the details under it, written from your side: "+10 points on Alice", then "Watching Valorant · Balance 12.4k points". Setup is in [docs/configuration.md](docs/configuration.md#notifications).

## Development

```sh
sh ci/check.sh                        # format, vet, build and test, as CI does
git config core.hooksPath ci/hooks   # run it before every push
```

The code is organised by domain under `internal/`. See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for the layout and conventions.

## Credits

- [0x8fv/Twitch-Channel-Points-Miner](https://github.com/0x8fv/Twitch-Channel-Points-Miner), the Go rewrite this project began as a fork of. This version reorganises the code and adds the dashboard, Telegram, notification sentences, the Docker image and CI.
- [rdavydov/Twitch-Channel-Points-Miner-v2](https://github.com/rdavydov/Twitch-Channel-Points-Miner-v2), the Python miner that inspired the rewrite.

## Licence and disclaimer

Released under [CC BY-NC 4.0](LICENSE), the licence of the project it derives from.

This project is not affiliated with Twitch and comes with no warranty. Automating an account can get it restricted or banned; you use it at your own risk.
