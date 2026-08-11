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

- Baize revision / Baize 版本：initial `custom/baize` baseline
- Legacy source / 旧代码来源：`main-v2@60d484837`
- Upstream WebUI changes / 上游 WebUI 变化：none in the reviewed range
- Ported / 已移植：all legacy fork net differences; `serve/web --dir`; single-row tool card header
- Intentionally skipped / 明确忽略：unrelated local `.gitignore` deletion
- Backend conflicts / 后端冲突：none during the squash import
- Verification / 验证：focused CLI and Serve tests; full gates recorded by the final migration commit
