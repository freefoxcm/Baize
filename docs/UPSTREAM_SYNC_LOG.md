# Upstream sync log / 上游同步日志

Every controlled upstream integration must add one entry before its merge commit.
每次受控上游整合都必须在 merge commit 前增加一条记录。

## Entry template / 记录模板

### YYYY-MM-DD — `<old-sha>` → `<new-sha>`

- Baize revision / Baize 版本：`<sha>`
- Upstream WebUI changes / 上游 WebUI 变化：`none` or summary
- Ported / 已移植：list or `none`
- Intentionally skipped / 明确忽略：list with reason or `none`
- Backend conflicts / 后端冲突：list and resolution or `none`
- Verification / 验证：commands and browser checks

## 2026-08-11 — `8a39ac1c4` → `ee2a6a766`

- Baize revision / Baize 版本：initial `custom/baize` migration series through `835465ec2`
- Legacy source / 旧代码来源：`main-v2@60d484837`
- Upstream WebUI changes / 上游 WebUI 变化：none in the reviewed range
- Ported / 已移植：all legacy fork net differences; `serve/web --dir`; single-row tool card header
- Intentionally skipped / 明确忽略：unrelated local `.gitignore` deletion
- Backend conflicts / 后端冲突：none during the squash import
- Verification / 验证：`git diff --check`; `node --check internal/serve/assets/baize.js`; focused CLI, Serve, i18n, Boot, hook, shell and built-in-tool tests; `go vet ./...`; `go run ./tools/repolint`; setup/sync script and pre-commit-hook integration tests; desktop dark-theme and 390×844 light-theme browser checks; static-resource headers and `/status` contract checks. `go test ./...` passed except Windows tests that require the OS symlink privilege (`internal/autoresearch`, `internal/installsource`, `internal/repair`, `internal/sessiontemp`). `make lint` was unavailable because `make` is not installed in the validation environment.

## 2026-08-11 — `ee2a6a766` → `bb58eec24`

- Baize revision / Baize 版本：pre-merge `custom/baize@8aa964b64`
- Branch migration / 分支迁移：archived legacy `main-v2@60d484837` as `legacy-main-v2-pre-baize-20260811`; changed the GitHub default to `custom/baize`; reset `main-v2` to the exact upstream SHA; removed `upstream-sync/main-v2`
- Upstream WebUI changes / 上游 WebUI 变化：none in Baize-owned Serve WebUI paths
- Ported / 已移植：the complete upstream range, including desktop context-capacity fixes and the native WebView2 approval smoke test
- Intentionally skipped / 明确忽略：none; upstream changes to `.github/workflows/ci.yml` and `.github/workflows/release-desktop.yml` remain tracked but inactive in this personal fork
- Backend conflicts / 后端冲突：none; the controlled merge completed cleanly
- Workflow state / Workflow 状态：the three repository-owned `baize-*` workflows are active and all 25 tracked upstream workflows are disabled; GitHub's platform-managed `dynamic/dependabot/update-graph` remains active to preserve dependency graph and vulnerability alerts
- Verification / 验证：`git diff --check`; fixed `actionlint v1.7.7`; `node --check internal/serve/assets/baize.js`; PowerShell and sh fork-maintenance tests; `go vet ./...`; `go run ./tools/repolint`; focused CLI, Serve, i18n, and Boot tests; CGO-disabled CLI build. `go test ./...` passed except the known Windows packages requiring symlink privilege (`internal/autoresearch`, `internal/installsource`, `internal/repair`, `internal/sessiontemp`); Linux Baize CI performs the complete suite.
