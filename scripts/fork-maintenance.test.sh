#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
test_root=$(mktemp -d "${TMPDIR:-/tmp}/baize-fork-test.XXXXXX")
trap 'rm -rf -- "$test_root"' EXIT

git init -q --bare "$test_root/upstream.git"
git init -q --bare "$test_root/origin.git"
git init -q -b main-v2 "$test_root/seed"
git -C "$test_root/seed" config user.name test
git -C "$test_root/seed" config user.email test@example.com
printf 'one\n' > "$test_root/seed/value.txt"
git -C "$test_root/seed" add value.txt
git -C "$test_root/seed" commit -qm one
git -C "$test_root/seed" remote add upstream "$test_root/upstream.git"
git -C "$test_root/seed" remote add origin "$test_root/origin.git"
git -C "$test_root/seed" push -q upstream main-v2
git -C "$test_root/seed" push -q origin main-v2

git clone -q -b main-v2 "$test_root/origin.git" "$test_root/work"
git -C "$test_root/work" config user.name test
git -C "$test_root/work" config user.email test@example.com
git -C "$test_root/work" remote add upstream "$test_root/upstream.git"
git -C "$test_root/work" switch -qc custom/baize
mkdir -p "$test_root/work/scripts"
cp "$repo_root/scripts/sync-upstream.sh" "$repo_root/scripts/baize-workflows-"*.txt "$test_root/work/scripts/"
cp "$repo_root/scripts/setup-fork-git.sh" "$test_root/work/scripts/"
mkdir -p "$test_root/work/.githooks"
cp "$repo_root/.githooks/pre-commit" "$repo_root/.githooks/pre-push" "$test_root/work/.githooks/"
git -C "$test_root/work" add scripts
git -C "$test_root/work" add .githooks
git -C "$test_root/work" commit -qm scripts
custom_before=$(git -C "$test_root/work" rev-parse custom/baize)
origin_before=$(git --git-dir="$test_root/origin.git" rev-parse main-v2)

printf 'two\n' >> "$test_root/seed/value.txt"
git -C "$test_root/seed" commit -qam two
git -C "$test_root/seed" push -q upstream main-v2
upstream_tip=$(git -C "$test_root/seed" rev-parse HEAD)

(cd "$test_root/work" && bash scripts/sync-upstream.sh >/dev/null)
[ "$(git -C "$test_root/work" branch --show-current)" = custom/baize ]
[ "$(git -C "$test_root/work" rev-parse main-v2)" = "$upstream_tip" ]
[ "$(git -C "$test_root/work" rev-parse custom/baize)" = "$custom_before" ]
[ "$(git --git-dir="$test_root/origin.git" rev-parse main-v2)" = "$origin_before" ]

git -C "$test_root/work" branch -m main-v2 main-v2-saved
if (cd "$test_root/work" && bash scripts/sync-upstream.sh >/dev/null 2>&1); then
  echo "sync accepted a missing main-v2 branch" >&2
  exit 1
fi
git -C "$test_root/work" branch -m main-v2-saved main-v2
git -C "$test_root/work" remote remove upstream
if (cd "$test_root/work" && bash scripts/sync-upstream.sh >/dev/null 2>&1); then
  echo "sync accepted a missing upstream remote" >&2
  exit 1
fi
git -C "$test_root/work" remote add upstream "$test_root/upstream.git"

# Git setup is idempotent and installs mirror-aware hooks.
(cd "$test_root/work" && sh scripts/setup-fork-git.sh >/dev/null)
(cd "$test_root/work" && sh scripts/setup-fork-git.sh >/dev/null)
[ "$(git -C "$test_root/work" config --local --get rerere.enabled)" = true ]
[ "$(git -C "$test_root/work" config --local --get pull.ff)" = only ]
[ "$(git -C "$test_root/work" config --local --get core.hooksPath)" = .git/baize-hooks ]
[ "$(git -C "$test_root/work" remote get-url --push upstream)" = DISABLED ]

hook_dir="$test_root/work/.git/baize-hooks"
(cd "$test_root/work" && sh "$hook_dir/pre-commit")
git -C "$test_root/work" switch -q main-v2
if (cd "$test_root/work" && sh "$hook_dir/pre-commit" >/dev/null 2>&1); then
  echo "pre-commit accepted a main-v2 commit" >&2
  exit 1
fi
git -C "$test_root/work" switch -q custom/baize

printf 'refs/heads/main-v2 %s refs/heads/main-v2 %s\n' "$upstream_tip" "$origin_before" |
  (cd "$test_root/work" && sh "$hook_dir/pre-push" origin ignored)
if printf 'refs/heads/main-v2 %s refs/heads/main-v2 %s\n' "$custom_before" "$origin_before" |
  (cd "$test_root/work" && sh "$hook_dir/pre-push" origin ignored >/dev/null 2>&1); then
  echo "pre-push accepted a non-mirror main-v2 SHA" >&2
  exit 1
fi
if printf 'refs/heads/main-v2 %040d refs/heads/main-v2 %s\n' 0 "$origin_before" |
  (cd "$test_root/work" && sh "$hook_dir/pre-push" origin ignored >/dev/null 2>&1); then
  echo "pre-push accepted deletion of main-v2" >&2
  exit 1
fi
if printf 'refs/heads/main-v2 %s refs/heads/main-v2 %s\n' "$upstream_tip" "$custom_before" |
  (cd "$test_root/work" && sh "$hook_dir/pre-push" origin ignored >/dev/null 2>&1); then
  echo "pre-push accepted a non-fast-forward main-v2 update" >&2
  exit 1
fi

# New upstream workflow files are fetched but require an explicit audit before
# the mirror may be published.
mkdir -p "$test_root/seed/.github/workflows"
printf 'name: unknown\n' > "$test_root/seed/.github/workflows/unknown.yml"
git -C "$test_root/seed" add .github/workflows/unknown.yml
git -C "$test_root/seed" commit -qm 'add unknown workflow'
git -C "$test_root/seed" push -q upstream main-v2
unknown_tip=$(git -C "$test_root/seed" rev-parse HEAD)
if (cd "$test_root/work" && bash scripts/sync-upstream.sh >/dev/null 2>&1); then
  echo "sync did not stop for an unknown upstream workflow" >&2
  exit 1
fi
[ "$(git -C "$test_root/work" branch --show-current)" = custom/baize ]
[ "$(git -C "$test_root/work" rev-parse main-v2)" = "$unknown_tip" ]

# A diverged mirror is rejected by --ff-only and the original branch is restored.
git -C "$test_root/work" switch -q main-v2
printf 'local\n' > "$test_root/work/local.txt"
git -C "$test_root/work" add local.txt
git -C "$test_root/work" commit --no-verify -qm local
git -C "$test_root/work" switch -q custom/baize
printf 'three\n' >> "$test_root/seed/value.txt"
git -C "$test_root/seed" commit -qam three
git -C "$test_root/seed" push -q upstream main-v2
if (cd "$test_root/work" && bash scripts/sync-upstream.sh >/dev/null 2>&1); then
  echo "sync accepted a diverged main-v2" >&2
  exit 1
fi
[ "$(git -C "$test_root/work" branch --show-current)" = custom/baize ]

touch "$test_root/work/dirty"
if (cd "$test_root/work" && bash scripts/sync-upstream.sh >/dev/null 2>&1); then
  echo "sync accepted a dirty worktree" >&2
  exit 1
fi
rm "$test_root/work/dirty"

mkdir -p "$test_root/github/scripts"
cp "$repo_root/scripts/setup-fork-github.sh" "$repo_root/scripts/baize-workflows-"*.txt "$test_root/github/scripts/"
git init -q -b custom/baize "$test_root/github"
git -C "$test_root/github" remote add origin https://github.com/freefoxcm/DeepSeek-Reasonix.git
cat > "$test_root/fake-gh" <<'FAKE'
#!/bin/sh
case "$1 $2" in
  "auth status") exit 0 ;;
  "repo view") printf 'custom/baize\n' ;;
  "workflow list")
    printf '%s\n' .github/workflows/baize-cache-impact.yml .github/workflows/baize-ci.yml .github/workflows/baize-docs-impact.yml
    ;;
  "workflow disable"|"workflow enable") exit 0 ;;
  "api repos/freefoxcm/DeepSeek-Reasonix")
    case "$*" in *permissions.admin*) printf 'true\n' ;; *) printf '{}\n' ;; esac
    ;;
  "api repos/freefoxcm/DeepSeek-Reasonix/actions/permissions") printf 'true\n' ;;
  api*) exit 0 ;;
  *) echo "unexpected fake gh invocation: $*" >&2; exit 1 ;;
esac
FAKE
chmod +x "$test_root/fake-gh"
(cd "$test_root/github" && BAIZE_GH_BIN="$test_root/fake-gh" sh scripts/setup-fork-github.sh --check >/dev/null)
(cd "$test_root/github" && BAIZE_GH_BIN="$test_root/fake-gh" sh scripts/setup-fork-github.sh --apply >/dev/null)

echo "fork maintenance sh tests passed"
