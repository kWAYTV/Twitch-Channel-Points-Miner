#!/bin/sh
# Starts the published image the way the README tells people to: compose.yaml,
# the example config and a ./data folder owned by the current user. Fails when
# the miner does not come up or cannot write to ./data.
set -u
cd "$(dirname "$0")/.."

mkdir -p data
cp config.example.json data/config.json
PUID=$(id -u)
PGID=$(id -g)
export PUID PGID

docker compose pull --quiet
docker compose up -d

started=1
for _ in 1 2 3 4 5 6 7 8 9 10; do
	if docker compose logs 2>&1 | grep -q "Dashboard listening on http://0.0.0.0:5000"; then
		started=0
		break
	fi
	sleep 0.5
done
docker compose logs 2>&1 | head -5
docker compose down

if [ "$started" -ne 0 ]; then
	echo "::error::the published image did not start with compose.yaml"
	exit 1
fi
if [ ! -d data/analytics ]; then
	echo "::error::the container could not write to ./data"
	exit 1
fi
echo "the image starts and writes to ./data"
