# Baize fork maintenance

This personal fork has two long-lived branches. These rules are mandatory for
people and coding agents working in the repository.

## Branch roles

| Branch | Role | Allowed changes |
|---|---|---|
| `main-v2` | Exact mirror of `upstream/main-v2` | Fast-forward updates only |
| `custom/baize` | Default development and deployment branch | Reviewed features and controlled upstream merges |
| `feat/*`, `fix/*` | Short-lived work based on `custom/baize` | One coherent change |

Never commit on `main-v2`, rebase or force-push `custom/baize`, push to
`upstream`, or mix feature work into an upstream merge commit. The pre-Baize
`main-v2` history is archived by the remote tag
`legacy-main-v2-pre-baize-20260811`.

## Configure a clone and GitHub

Run the local setup once per clone:

```powershell
pwsh -NoProfile -File scripts/setup-fork-git.ps1
```

```sh
sh scripts/setup-fork-git.sh
```

It enables rerere, fast-forward-only pulls, pruning, persistent hooks, the
Baize merge driver, and a fetch-only `upstream` remote. Check the GitHub-side
configuration with `scripts/setup-fork-github --check`; repository admins may
apply the expected default branch, workflow allowlist, and branch protections
with `--apply`.

Only these workflows are active in this personal fork:

- `.github/workflows/baize-ci.yml`
- `.github/workflows/baize-cache-impact.yml`
- `.github/workflows/baize-docs-impact.yml`

Upstream workflow files remain tracked to reduce merge conflicts, but their
repository workflow state is disabled. Official releases, signing, npm,
Cloudflare, Pages, E2E bots, labels, scheduled maintenance, and community
automation are inactive. GitHub may also display the platform-managed
`dynamic/dependabot/update-graph`; it is not a repository-owned workflow and is
excluded from the Baize allowlist check because disabling it would also disable
the repository's dependency graph and vulnerability alerts.

## Fetch and integrate upstream

Update only the mirror:

```powershell
pwsh -NoProfile -File scripts/sync-upstream.ps1
```

```sh
sh scripts/sync-upstream.sh
```

The script requires a clean worktree, returns to the original branch, verifies
that local `main-v2` exactly matches `upstream/main-v2`, and never pushes or
merges `custom/baize`. Review its overall, WebUI, and workflow summaries before
publishing the mirror:

```sh
git push origin main-v2
git switch custom/baize
git merge --no-ff --no-commit main-v2
```

If upstream adds an unknown workflow, stop. Temporarily disable repository
Actions for the first mirror/custom push containing that file, add the path to
the disabled inventory, apply the GitHub setup, and only then re-enable Actions.

The Baize merge driver keeps fork-owned frontend files. Resolve backend
conflicts normally, manually port wanted upstream WebUI changes, and update
`docs/UPSTREAM_SYNC_LOG.md`. Use `git merge --abort` before committing an
unready integration; revert an already shared merge instead of rewriting it.

## WebUI ownership and verification

Baize owns the paths marked `merge=baize` in `.gitattributes`: main/login HTML,
logos, `assets/baize.css`, and `assets/baize.js`. Keep structure, presentation,
and behavior in HTML, CSS, and JavaScript respectively. Backend routes and APIs
always use normal merges and focused Go tests.

Use Conventional Commits. Before feature commits run focused tests, gofmt, and
`git diff --check`. Before an upstream integration run `go vet ./...`, repolint,
`go test ./...`, and browser checks for WebUI changes. Record old/new upstream
SHAs, ported/skipped UI work, conflicts, and verification in the sync log.
