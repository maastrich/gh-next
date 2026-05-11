# gh-next

> What needs your attention on GitHub — PRs, issues, and discussions, sorted by whose turn it is.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/maastrich/gh-next/main/install.sh | bash
```

Checks prerequisites, installs the extension, sets up a recurring schedule (weekdays 8am–6pm, hourly), and runs the first fetch.

**Manual install:**
```sh
gh extension install maastrich/gh-next
```

## Commands

```sh
gh next                              # Show cached status
gh next status                       # Fetch fresh data and show status
gh next explore <path>               # Index all git repos under <path>
gh next program                      # Schedule hourly runs (weekdays 8am–6pm)
gh next program "<cron>"             # Custom cron expression
gh next program --show               # Show current schedule
gh next program --remove             # Remove the cron job
```

### `gh next status`

Fetches all open PRs (authored + review-requested), open issues, and open discussions via the GitHub GraphQL API, classifies them, and displays them grouped by who needs to act.

```sh
gh next status
gh next status --stale-days 14
```

### `gh next program [cron-expression]`

Installs a cron job that runs `gh next status` on a schedule. Default: every hour from 8am to 6pm, Monday–Friday (`0 8-18 * * 1-5`), in your system timezone.

```sh
gh next program                     # default weekday schedule
gh next program "0 * * * *"         # every hour, all day
gh next program "0 9-17 * * 1-5"    # 9am–5pm weekdays
gh next program --show              # show active schedule
gh next program --remove            # remove the job
```

### `gh next explore <path>`

Walks the directory tree, finds every git repository with a GitHub remote, and writes `~/.gh-next/gh-index.json` mapping GitHub URLs to local filesystem paths. Handles multiple clones of the same repo.

```sh
gh next explore ~/code
gh next explore ~/contributions
```

## Status categories

| Group | Status | Meaning |
|-------|--------|---------|
| 🟢 Your turn | ⚔️ merge_conflict | Needs rebase |
| | ✅ ready_to_merge | Approved + CI green, you can merge |
| | 🔧 changes_needed | Changes requested or CI failing |
| | 💬 awaiting_answer | Reviewer commented since last push |
| | 💬 needs_reply | Someone replied to your issue/discussion |
| | 👀 review_requested | Someone's PR waiting for your review |
| | 🔁 re_review_requested | You reviewed, they pushed new commits |
| 🟡 Their turn | ⏳ awaiting_merge | Approved, waiting on maintainer |
| | ⚙️ ci_running | Tests in progress |
| | 🔍 awaiting_review | No review yet |
| | 🔍 awaiting_response | Waiting for reply on issue/discussion |
| ⚪ Parked | 📝 draft | Draft PR |
| | ✅ answered | Discussion marked answered |
| | 🕸️ stale | No activity in 7+ days |

## Prerequisites

| Tool | Required | Install |
|------|----------|---------|
| [`gh`](https://cli.github.com/) | ✅ | `brew install gh` + `gh auth login` |
| [`jq`](https://jqlang.github.io/jq/) | ✅ | `brew install jq` |
| [`terminal-notifier`](https://github.com/julienXX/terminal-notifier) | ✅ | `brew install terminal-notifier` |

The curl install script handles all of these automatically.
