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

## 2026-08-13 — `bb58eec24` → `42a9de71d`

- Baize revision / Baize 版本：pre-merge `custom/baize@790925f3e`
- Upstream WebUI changes / 上游 WebUI 变化：desktop gained native transcript virtualization, recovery-lineage presentation, indexed history search, multi-currency billing, task/session catalogs, search citations, and refreshed execution settings; the Baize-owned Serve UI received the applicable billing contract manually instead of being replaced
- Ported / 已移植：the complete upstream kernel through `42a9de71d`, including unified execution roles, goal budgets and verification termination, session write authority and recovery lineage, projection catalogs, command-effect classification, compaction state, plugin registration safety, provider URL/billing/search contracts, and the final inbox isolation, encrypted-search accounting, and Auto Guard commits; Baize Serve now consumes `costQuote`/`sessionCostQuote`, renders single- and multi-currency totals, and keeps the existing Baize identity, workspace selection, effort fallback, token/tool-duration accounting, plan-mode semantics, steering guards, and request-independent controller lifecycle
- Baize WebUI fixes / Baize WebUI 修复：PDF workspace previews keep the authenticated same-origin content endpoint but no longer sandbox Chrome's built-in PDF viewer; HTML and SVG isolation remains unchanged; Windows Serve tests now fence shared history and usage projections before temporary profile cleanup
- Intentionally skipped / 明确忽略：desktop-only React Virtuoso, native window diagnostics, and desktop theme presentation were not duplicated into the Baize Serve HTML/CSS/JS surface; those upstream desktop implementations remain present in their own runtime
- Backend conflicts / 后端冲突：resolved `.github/workflows/ci.yml`, Boot golden data, `internal/cli/cli.go`, `internal/provider/provider.go`, `internal/serve/serve.go`, and the repolint baseline; retained Baize CI timeouts and third-party WebView tests, preserved provider server-search and response items with Baize tool durations, handled `SetSessionLeases` errors explicitly, and combined upstream session write authority with Baize's request-independent controller rebuild context
- Workflow state / Workflow 状态：no unknown workflow was introduced; the known upstream `ci.yml` and `deploy-crash-worker.yml` changes were audited, and `scripts/setup-fork-github.ps1 --check` confirmed only the three `baize-*` workflows plus GitHub's dynamic Dependabot graph workflow remain active
- Verification / 验证：`git diff --check`; `node --check internal/serve/assets/baize.js`; `go run ./tools/repolint`; focused Agent, CLI, Control, Boot, Provider, Skill, ShellSafe, Serve, Session/History/Usage/Task Catalog tests; PDF inline/range and no-sandbox regression tests; live in-app Chromium rendered the full PDF without console errors or a blocked-page interstitial; `go vet ./...`; `go test ./...` passed all code-related packages, with only environment-bound POSIX Bash and Windows symlink-privilege tests unavailable; `go build -o reasonix.exe ./cmd/reasonix`; live IPAP capability discovery and MCP/Skill smoke checks

## 2026-08-13 — `42a9de71d` → `9aaf8d381`

- Baize revision / Baize 版本：pre-merge `custom/baize@bda89c4b4`
- Upstream WebUI changes / 上游 WebUI 变化：desktop-only project-tree recovery presentation; no Baize Serve WebUI path changed
- Ported / 已移植：the complete one-commit follow-up, adding canonical session retargeting and indexed session-directory scanning so upgraded conversations remain visible and reopen at the latest recovery leaf
- Intentionally skipped / 明确忽略：no feature commit was skipped; desktop presentation stays in the desktop surface and was not duplicated into Baize Serve
- Backend conflicts / 后端冲突：none; the incremental controlled merge completed cleanly
- Workflow state / Workflow 状态：no workflow file changed or was added
- Verification / 验证：`git diff --check`; `go run ./tools/repolint`; Session Catalog and focused desktop tests; `go vet ./...`; `go test ./...`; Baize Linux, Windows, cache-impact and docs-impact CI

## 2026-08-13 — `9aaf8d381` → `c7bc2f3e1`

- Baize revision / Baize 版本：pre-merge `custom/baize@d6e246c47`
- Upstream WebUI changes / 上游 WebUI 变化：desktop-only project-tree status label and styling; no Baize Serve WebUI path changed
- Ported / 已移植：the complete one-commit follow-up distinguishing delivery-check pauses from recovery pauses in desktop session runtime status
- Intentionally skipped / 明确忽略：none; no duplicate Serve presentation was needed
- Backend conflicts / 后端冲突：none; the incremental merge completed cleanly
- Workflow state / Workflow 状态：no workflow file changed or was added
- Verification / 验证：`git diff --check`; `go run ./tools/repolint`; focused desktop runtime-status tests; `go vet ./...`; Baize Linux, Windows, cache-impact and docs-impact CI

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
