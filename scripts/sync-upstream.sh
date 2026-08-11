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
git show-ref --verify --quiet refs/heads/upstream-sync/main-v2 || {
  echo "error: local branch upstream-sync/main-v2 does not exist" >&2
  exit 1
}
git remote get-url upstream >/dev/null

old_revision=$(git rev-parse upstream-sync/main-v2)
switched=0
restore_branch() {
  if [ "$switched" -eq 1 ]; then
    git switch "$current_branch" >/dev/null 2>&1 || true
  fi
}
trap restore_branch EXIT HUP INT TERM

git fetch upstream --prune
if [ "$current_branch" != "upstream-sync/main-v2" ]; then
  git switch upstream-sync/main-v2
  switched=1
fi
git merge --ff-only upstream/main-v2
new_revision=$(git rev-parse upstream-sync/main-v2)
if [ "$switched" -eq 1 ]; then
  git switch "$current_branch"
  switched=0
fi
trap - EXIT HUP INT TERM

echo "upstream-sync/main-v2: $old_revision -> $new_revision"
if [ "$old_revision" != "$new_revision" ]; then
  git diff --stat "$old_revision..$new_revision" -- \
    internal/serve/index.html internal/serve/login.html internal/serve/serve.go
fi
echo "No merge or push was performed. Next, inspect the diff and run:"
echo "  git switch custom/baize"
echo "  git merge --no-ff --no-commit upstream-sync/main-v2"
