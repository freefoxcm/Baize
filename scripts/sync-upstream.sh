#!/bin/sh
set -eu

repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || {
  echo "error: run this script from inside the Reasonix repository" >&2
  exit 1
}
cd "$repo_root"

current_branch=$(git symbolic-ref --quiet --short HEAD 2>/dev/null) || {
  echo "error: detached HEAD is not supported" >&2
  exit 1
}
if [ -n "$(git status --porcelain)" ]; then
  echo "error: working tree must be clean before syncing upstream" >&2
  exit 1
fi
for state in MERGE_HEAD rebase-merge rebase-apply; do
  if [ -e "$(git rev-parse --git-path "$state")" ]; then
    echo "error: finish or abort the current merge/rebase before syncing upstream" >&2
    exit 1
  fi
done
git show-ref --verify --quiet refs/heads/main-v2 || {
  echo "error: local branch main-v2 does not exist" >&2
  exit 1
}
git remote get-url upstream >/dev/null

old_revision=$(git rev-parse main-v2)
switched=0
restore_branch() {
  if [ "$switched" -eq 1 ]; then
    git switch "$current_branch" >/dev/null 2>&1 || true
  fi
}
trap restore_branch EXIT HUP INT TERM

git fetch upstream --prune
if [ "$current_branch" != "main-v2" ]; then
  git switch main-v2
  switched=1
fi
git merge --ff-only upstream/main-v2
new_revision=$(git rev-parse main-v2)
upstream_revision=$(git rev-parse upstream/main-v2)
main_tree=$(git rev-parse 'main-v2^{tree}')
upstream_tree=$(git rev-parse 'upstream/main-v2^{tree}')
if [ "$new_revision" != "$upstream_revision" ] || [ "$main_tree" != "$upstream_tree" ]; then
  echo "error: main-v2 does not exactly match upstream/main-v2 after synchronization" >&2
  exit 1
fi
if [ "$switched" -eq 1 ]; then
  git switch "$current_branch"
  switched=0
fi
trap - EXIT HUP INT TERM

echo "main-v2: $old_revision -> $new_revision"
if [ "$old_revision" != "$new_revision" ]; then
  echo "Overall upstream changes:"
  git diff --stat "$old_revision..$new_revision"
  echo "Upstream WebUI changes requiring Baize review:"
  git diff --stat "$old_revision..$new_revision" -- \
    internal/serve/index.html internal/serve/login.html internal/serve/serve.go \
    internal/serve/web_assets.go internal/serve/assets
  echo "Upstream workflow changes:"
  git diff --stat "$old_revision..$new_revision" -- .github/workflows
fi

unknown=""
if [ "$old_revision" != "$new_revision" ]; then
  for path in $(git diff --diff-filter=A --name-only "$old_revision..$new_revision" -- .github/workflows); do
    if ! grep -Fqx "$path" scripts/baize-workflows-active.txt && \
       ! grep -Fqx "$path" scripts/baize-workflows-disabled.txt; then
      unknown="${unknown}${unknown:+
}$path"
    fi
  done
fi
if [ -n "$unknown" ]; then
  echo "error: new upstream workflows require an Actions-disabled audit before pushing main-v2:" >&2
  printf '  %s\n' "$unknown" >&2
  exit 1
fi

echo "No push or custom merge was performed. Next, inspect the diff and run:"
echo "  git push origin main-v2"
echo "  git switch custom/baize"
echo "  git merge --no-ff --no-commit main-v2"
