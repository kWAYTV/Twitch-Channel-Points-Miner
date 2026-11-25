# Getting started

## Docker

You need Docker with the Compose plugin. Nothing is built locally.

1. Download [compose.yaml](../compose.yaml) and [config.example.json](../config.example.json) into an empty folder.
2. Create the data folder and the config:

   ```sh
   mkdir data
   cp config.example.json data/config.json
   ```

3. Set `username` in `data/config.json` to your Twitch login.
4. Start it and follow the logs:

   ```sh
   docker compose up -d
   docker compose logs -f
   ```

5. [Log in](#logging-in) with the code the logs show.

The image is `ghcr.io/kwaytv/twitch-points-miner:latest`, for `linux/amd64` and `linux/arm64`. Each build is also tagged `sha-<commit>` if you want to pin a version.

### User and permissions

The container runs as user 1000 and writes to `./data`. If your user id is different (check with `id -u`), pass it in:

```sh
PUID=$(id -u) PGID=$(id -g) docker compose up -d
```

or put `PUID` and `PGID` in a `.env` file next to `compose.yaml`. `TZ` sets the time zone of the timestamps the same way, for example `TZ=Europe/Madrid`.

### Without Compose

```sh
docker run -d --name twitch-miner --restart unless-stopped \
  -v "$PWD/data:/data" -p 5000:5000 \
  ghcr.io/kwaytv/twitch-points-miner:latest
```

If `data/config.json` does not exist, the first start writes one with defaults. Set `username` in it and restart the container.

## From source

Requires Go 1.23 or newer.

```sh
git clone https://github.com/kWAYTV/twitch-points-miner
cd twitch-points-miner
go run ./cmd/twitch-miner
```

The first run creates `config.json`. Set `username`, run it again and [log in](#logging-in). To build a binary instead, run `go build ./cmd/twitch-miner`.

Press Ctrl+C to stop. A summary of the session is printed.

## Logging in

The miner uses Twitch's device login, so it never asks for your password. On the first start it prints:

```
Open https://www.twitch.tv/activate and enter code: ABCDEFGH (expires in 30 minutes)
```

Open the page in a browser where you are signed in to Twitch and enter the code. The login is saved in `cookies/<username>.json` and reused on later starts. To log in again, or with another account, delete that file and restart.

## Where files are kept

Everything lives in one folder, next to `config.json`:

| Path | Contents |
|---|---|
| `config.json` | Your settings. |
| `cookies/<username>.json` | The saved Twitch login. Keep it private. |
| `analytics/` | The dashboard's history, one set of files per streamer. |
| `log/` | The log file when `save_logs` is on, and the watch-streak cache. |

The folder is chosen in this order:

1. `-data-dir <folder>`, or `-config <file>` (the folder of that file).
2. `TCPM_DATA_DIR`, or `TCPM_CONFIG`.
3. The current directory, if it has a `config.json`.
4. The directory of the executable.

In the Docker image it is `/data`.

## Updating

- **Docker:** `docker compose pull && docker compose up -d`.
- **From source:** `git pull`, then run or build again.

A binary can replace itself when a newer release is published (`auto_update`). This repository has not published a release yet, so the check finds nothing and does nothing. It is always off in the Docker image.

Your `config.json` carries over. A key added by a newer version is filled in with its default on the next start.
