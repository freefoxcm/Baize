#!/bin/sh
set -eu

repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || {
  echo "error: run this script from inside the Reasonix repository" >&2
  exit 1
}
cd "$repo_root"

git remote get-url upstream >/dev/null
git_dir=$(git rev-parse --git-dir)
hook_dir="$git_dir/baize-hooks"
mkdir -p "$hook_dir"
cp .githooks/pre-commit .githooks/pre-push "$hook_dir/"
chmod +x "$hook_dir/pre-commit" "$hook_dir/pre-push"

git config --local rerere.enabled true
git config --local rerere.autoupdate true
git config --local pull.ff only
git config --local fetch.prune true
git config --local core.hooksPath "$hook_dir"
git config --local merge.baize.name "Keep Baize-owned frontend during controlled upstream merges"
git config --local merge.baize.driver true
git remote set-url --push upstream DISABLED

echo "Baize fork Git settings installed for $repo_root"
echo "upstream is fetch-only; custom/baize remains the integration branch"
