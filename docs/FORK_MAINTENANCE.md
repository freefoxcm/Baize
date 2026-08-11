# Baize fork maintenance

This fork separates upstream intake from product development. These rules are
mandatory for people and coding agents working in this repository.

## Branch roles

| Branch | Role | Allowed changes |
|---|---|---|
| `upstream-sync/main-v2` | Clean mirror of `upstream/main-v2` | Fast-forward updates only |
| `custom/baize` | Baize integration and deployment | Reviewed feature commits and controlled upstream merges |
| `feat/*`, `fix/*` | Short-lived work | One coherent change based on `custom/baize` |
| `main-v2` | Preserved legacy history | No new Baize development |

Never commit on the sync branch, rebase `custom/baize`, push to `upstream`, or
mix feature work into an upstream merge commit.

## Configure a clone

Run one setup script from the repository root:

```powershell
pwsh -NoProfile -File scripts/setup-fork-git.ps1
```

```sh
sh scripts/setup-fork-git.sh
```

The script enables `rerere`, fast-forward-only pulls, pruning, persistent hooks
under `.git/baize-hooks`, the Baize merge driver, and a fetch-only `upstream`
remote. It is idempotent and keeps branch protection active on the clean mirror.

## Fetch and integrate upstream

First update only the clean mirror:

```powershell
pwsh -NoProfile -File scripts/sync-upstream.ps1
```

```sh
sh scripts/sync-upstream.sh
```

The script requires a clean worktree, returns to the original branch, and never
merges, commits, or pushes. Review the reported range, especially upstream's
frontend changes:

```sh
old_base=$(git merge-base custom/baize upstream-sync/main-v2)
git diff "$old_base..upstream-sync/main-v2" -- \
  internal/serve/index.html internal/serve/login.html internal/serve/serve.go
```

Then start a controlled merge:

```sh
git switch custom/baize
git merge --no-ff --no-commit upstream-sync/main-v2
```

The configured `baize` driver keeps fork-owned frontend files. Resolve backend
conflicts normally, manually port any wanted upstream WebUI changes into the
Baize assets, and update `docs/UPSTREAM_SYNC_LOG.md`. Run the required checks
before committing the merge and pushing both branches to `origin`.

Use `git merge --abort` before the merge commit if the integration is not ready.
After a committed integration, use a normal revert commit; do not rewrite the
shared `custom/baize` history.

## WebUI ownership

Baize owns the paths marked `merge=baize` in `.gitattributes`: the main and login
HTML, Baize logos, `assets/baize.css`, and `assets/baize.js`. Keep HTML structural,
put presentation in the CSS asset, and put browser behavior in the JavaScript
asset. Keep the theme bootstrap inline only to prevent first-paint flashing.

Backend routes and APIs are not fork-protected. Add new Serve behavior in focused
Go files with tests, and let upstream merge normally. Never mark backend code
with the Baize merge driver to hide a conflict.

## Commit and verification rules

- Use Conventional Commits and keep each feature, fix, WebUI change, upstream
  integration, and documentation change independently reviewable.
- Before a feature commit, run focused tests plus `gofmt` and `git diff --check`.
- Before an upstream merge or push, run `go vet ./...`, `make lint`, and
  `go test ./...`, plus browser checks when the WebUI changed.
- Record old/new upstream SHAs, ported and skipped UI work, and verification
  results in the sync log. A merge is incomplete until the log is updated.
