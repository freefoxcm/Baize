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
