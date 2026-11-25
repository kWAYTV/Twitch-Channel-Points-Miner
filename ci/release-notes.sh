#!/bin/sh
# Prints the notes of a release: the changes grouped by kind, how to install
# it, and a link to everything since the previous release.
#
#   NOTES    the changelog the version step generated
#   TAG      the new tag, such as v1.6.1
#   VERSION  the same without the v
#
# Run before the new tag is fetched, so the latest local tag is the previous release.
set -eu

previous=$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)
image=ghcr.io/$(printf '%s' "$GITHUB_REPOSITORY" | tr '[:upper:]' '[:lower:]')

# Without the generated title line: the release already carries the version.
printf '%s\n' "$NOTES" | sed '1{/^##* \[/d}' | cat -s

cat <<EOF

### Install

- Docker: \`$image:$VERSION\`
- Binaries for Windows, Linux and macOS are attached below. A running binary with \`auto_update\` on installs this release by itself.
EOF

if [ -n "$previous" ]; then
	printf '\n[All changes since %s](https://github.com/%s/compare/%s...%s)\n' "$previous" "$GITHUB_REPOSITORY" "$previous" "$TAG"
fi
