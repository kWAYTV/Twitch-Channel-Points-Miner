# Contributing to this repository

Thank you for your interest in contributing to **Twitch-Channel-Points-Miner**!  
This project is community-driven, and every improvement—big or small—helps make the miner more reliable and feature-rich.

---

## Getting started

Before contributing:

- Please review the **[Code of Conduct](./CODE_OF_CONDUCT.md)**.
- Check **[existing issues](https://github.com/kWAYTV/twitch-points-miner/issues)** to avoid duplicates.

If an issue related to your idea already exists, join the discussion instead of opening a new one.

---

## Ready to make a change? Fork the repo

### Fork using GitHub Desktop
- Install & set up GitHub Desktop:  
  <https://docs.github.com/en/desktop/installing-and-configuring-github-desktop/getting-started-with-github-desktop>
- Learn how to clone and fork repositories using Desktop:  
  <https://docs.github.com/en/desktop/contributing-and-collaborating-using-github-desktop/cloning-and-forking-repositories-from-github-desktop>

### Fork using the command line
Guide on how to fork a repo:  
<https://docs.github.com/en/get-started/quickstart/fork-a-repo>

### Fork using GitHub Codespaces
Run and edit the project in the cloud with no local setup required:  
<https://docs.github.com/en/codespaces/developing-in-codespaces/creating-a-codespace>

---

## Opening a Pull Request (PR)

Once your changes are ready, open a **[Pull Request](https://github.com/kWAYTV/twitch-points-miner/pulls)**.  
Be sure to fill out the PR template so reviewers understand your changes and intent.

---

## Getting your PR reviewed

- Begin with a **self-review** (checklist below).
- Maintainers/community contributors will review your PR and may request changes.
- Monitor your PR for comments and questions.
- If you encounter merge conflicts, see:  
  **[Resolving merge conflicts](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/addressing-merge-conflicts)**

### After your PR is merged
🎉 Congratulations — you’re now a contributor!  
You’ll appear in the **[Contributor Graph](https://github.com/kWAYTV/twitch-points-miner/graphs/contributors)**.

---

# Types of contributions 📝

You can help by:

- Reporting bugs  
- Improving documentation  
- Submitting patches or new features  
- Suggesting enhancements  
- Writing or improving tests  
- Helping with code quality and performance  

Every contribution is appreciated!

---

## Issues

Use **[Issues](https://github.com/kWAYTV/twitch-points-miner/issues)** to report:

- Bugs  
- Feature requests  
- Questions  
- Enhancement ideas  
- Clarifications  

Please use the appropriate template when opening an issue.

---

## Labels

Labels help identify an issue’s purpose and status:

- **`bug`** — Something isn’t working  
- **`documentation`** — README or docs updates  
- **`duplicate`** — Already reported  
- **`enhancement`** — New feature or improvement  
- **`help wanted`** — Help needed from contributors  
- **`improvements`** — Improve existing feature  
- **`invalid`** — Issue doesn’t meet guidelines  
- **`question`** — Additional details needed  
- **`wontfix`** — Issue won’t be addressed  

Browse all labels here:  
<https://github.com/kWAYTV/twitch-points-miner/labels>

---

# Opening a Pull Request

For small edits (documentation, typos), feel free to use GitHub’s web editor.

For code changes:

1. **Fork** the repository  
2. **Clone** your fork  
3. Create a **new branch**  
4. Implement your updates  
5. Run formatting, linting, vetting, and tests  
6. Commit & push  
7. Open a **Pull Request**  

General PR documentation:  
<https://docs.github.com/en/pull-requests>

---

# Go Style Guide

This project follows standard Go tooling and conventions.

---

## Required Formatting

All Go code **must** follow `gofmt` formatting.

Run:

```bash
gofmt -w .
```

Or:

```bash
go fmt ./...
```

Unformatted code will not be accepted.

---

## Project Layout

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
| `internal/notify` | Discord and Telegram notifications and the sentences they send. |
| `internal/dashboard` | The analytics store, its HTTP API and the web page. |
| `internal/project` | The project's name, version and repository. |
| `internal/privacy`, `internal/update` | Log anonymisation; the self-updater. |
| `internal/ansi`, `internal/jsonfile`, `internal/jsonvalue`, `internal/random` | Console colours; writing JSON files; reading loose JSON values; nonces and jitter. |

Keep files small and single-purpose, and functions short. Put new code in the package that owns its subject; add a package when a subject has none.

---

## Checks and Git Hooks

CI runs three jobs side by side: formatting, vetting and module tidiness; the tests; and a build for every released platform. The same checks run locally with one script:

```bash
sh ci/check.sh        # everything
sh ci/check.sh lint   # gofmt, go vet, go mod tidy
sh ci/check.sh test   # tests
sh ci/check.sh build  # every released platform
```

Enable the hooks once per clone so they run without being remembered:

```bash
git config core.hooksPath .githooks
```

- `pre-commit` formats and vets the Go files being committed.
- `pre-push` runs `ci/check.sh`, the same checks CI runs.

Rules for CI changes: never retry a failing check, never skip or loosen one, and never raise a time budget without a measured reason.

---

## Linting

We recommend using **golangci-lint**:

- Documentation: <https://golangci-lint.run/>

Run linting:

```bash
golangci-lint run
```

---

## Vetting

Run Go’s vet tool to catch common mistakes:

```bash
go vet ./...
```

---

## Testing

Use tests to ensure functionality remains stable:

```bash
go test ./...
```

All tests must pass before your PR will be accepted.

---

# Self-review checklist

Before marking your PR as “Ready for review,” ensure:

- [ ] The PR fully addresses the linked issue  
- [ ] Code compiles with no errors  
- [ ] All Go files are formatted (`gofmt` / `go fmt`)  
- [ ] `go vet` shows no critical issues  
- [ ] Linting (`golangci-lint run`) passes  
- [ ] All tests pass (`go test ./...`)  
- [ ] Code follows Go best practices and project structure  
- [ ] Documentation is updated where needed  
- [ ] All CI checks are passing  

---

## Suggested changes from reviewers

Reviewers may request modifications using:

- Inline comments  
- Suggested code snippets  
- Review summaries  

When updating your PR:

- Push commits normally (do **not** squash unless asked)
- Respond to reviewer feedback
- Mark discussions as **[Resolved](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests#resolving-conversations)** using GitHub’s interface

---

Thank you for contributing to **Twitch-Channel-Points-Miner**! 🚀  
Your help makes the project better for everyone.
