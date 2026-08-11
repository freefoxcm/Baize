#!/bin/sh
set -eu

mode=check
case "${1:-}" in
  ""|--check) mode=check ;;
  --apply) mode=apply ;;
  *) echo "usage: scripts/setup-fork-github.sh [--check|--apply]" >&2; exit 2 ;;
esac
gh_bin=${BAIZE_GH_BIN:-gh}
run_gh() { "$gh_bin" "$@"; }

repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || {
  echo "error: run this script from inside the Reasonix repository" >&2
  exit 1
}
cd "$repo_root"

origin=$(git remote get-url origin)
repository=$(printf '%s\n' "$origin" | sed -E 's#^https://github\.com/##; s#^git@github\.com:##; s#\.git$##')
case "$repository" in
  */*) ;;
  *) echo "error: origin is not a supported GitHub URL: $origin" >&2; exit 1 ;;
esac
if [ "$(printf '%s' "$repository" | tr '[:upper:]' '[:lower:]')" = "esengine/deepseek-reasonix" ]; then
  echo "error: refusing to configure the upstream repository" >&2
  exit 1
fi

run_gh auth status >/dev/null
[ "$(run_gh api "repos/$repository" --jq '.permissions.admin')" = "true" ] || {
  echo "error: GitHub admin permission is required for $repository" >&2
  exit 1
}

if [ "$mode" = apply ]; then
  workflow_states=$(run_gh workflow list --repo "$repository" --all --json path,state \
    --jq '.[] | [.path, .state] | @tsv')
  while IFS= read -r path; do
    [ -n "$path" ] || continue
    if printf '%s\n' "$workflow_states" | grep -Fqx "$(printf '%s\tactive' "$path")"; then
      run_gh workflow disable "$path" --repo "$repository"
    fi
  done < scripts/baize-workflows-disabled.txt

  run_gh api --method PATCH "repos/$repository" -f default_branch=custom/baize >/dev/null

  protection='{"required_status_checks":null,"enforce_admins":false,"required_pull_request_reviews":null,"restrictions":null,"required_linear_history":false,"allow_force_pushes":false,"allow_deletions":false,"required_conversation_resolution":false}'
  for branch in main-v2 custom/baize; do
    encoded=$(printf '%s' "$branch" | sed 's#/#%2F#g')
    printf '%s' "$protection" | run_gh api --method PUT "repos/$repository/branches/$encoded/protection" --input - >/dev/null
  done

  while IFS= read -r path; do
    [ -n "$path" ] || continue
    attempt=1
    until run_gh workflow enable "$path" --repo "$repository"; do
      [ "$attempt" -lt 5 ] || { echo "error: failed to enable $path" >&2; exit 1; }
      attempt=$((attempt + 1))
      sleep 2
    done
  done < scripts/baize-workflows-active.txt
fi

problems=0
default_branch=$(run_gh repo view "$repository" --json defaultBranchRef --jq '.defaultBranchRef.name')
if [ "$default_branch" != custom/baize ]; then
  echo "error: default branch is $default_branch (expected custom/baize)" >&2
  problems=$((problems + 1))
fi
if [ "$(run_gh api "repos/$repository/actions/permissions" --jq '.enabled')" != true ]; then
  echo "error: GitHub Actions is disabled for the repository" >&2
  problems=$((problems + 1))
fi

active=$(run_gh workflow list --repo "$repository" --all --json path,state --jq '.[] | select(.state == "active") | .path')
printf '%s\n' "$active" | while IFS= read -r path; do
  [ -z "$path" ] || grep -Fqx "$path" scripts/baize-workflows-active.txt || {
    echo "error: unexpected active workflow: $path" >&2
    exit 17
  }
done || problems=$((problems + 1))
while IFS= read -r path; do
  [ -n "$path" ] || continue
  printf '%s\n' "$active" | grep -Fqx "$path" || {
    echo "error: Baize workflow is not active: $path" >&2
    problems=$((problems + 1))
  }
done < scripts/baize-workflows-active.txt
for branch in main-v2 custom/baize; do
  encoded=$(printf '%s' "$branch" | sed 's#/#%2F#g')
  run_gh api "repos/$repository/branches/$encoded/protection" --silent || {
    echo "error: $branch is not protected" >&2
    problems=$((problems + 1))
  }
done

echo "repository: $repository"
echo "default branch: $default_branch"
echo "active workflows:"
printf '%s\n' "$active" | sed '/^$/d; s/^/  /'
[ "$problems" -eq 0 ] || exit 1
echo "Baize GitHub fork configuration is valid"
