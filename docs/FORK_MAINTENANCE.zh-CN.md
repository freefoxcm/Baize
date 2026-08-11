# Baize 分叉维护规范

本个人分叉只保留两个长期分支。以下规则同时约束开发者和编码代理。

## 分支职责

| 分支 | 职责 | 允许的改动 |
|---|---|---|
| `main-v2` | `upstream/main-v2` 的精确镜像 | 只能快进更新 |
| `custom/baize` | 默认开发和部署分支 | 已审查的功能与受控上游合并 |
| `feat/*`、`fix/*` | 基于 `custom/baize` 的短期开发 | 一个完整、单一职责的改动 |

禁止在 `main-v2` 提交，禁止 rebase 或强推 `custom/baize`，禁止向
`upstream` 推送，也禁止把功能开发混入上游 merge commit。旧 `main-v2`
历史由远程 tag `legacy-main-v2-pre-baize-20260811` 永久归档。

## 配置克隆与 GitHub

每个克隆运行一次本地设置：

```powershell
pwsh -NoProfile -File scripts/setup-fork-git.ps1
```

```sh
sh scripts/setup-fork-git.sh
```

它会启用 rerere、仅快进 pull、远程清理、持久 hooks、Baize merge driver，
并禁用向 `upstream` 推送。使用 `scripts/setup-fork-github --check` 检查
GitHub 侧状态；管理员可用 `--apply` 幂等设置默认分支、Workflow allowlist
和分支保护。

本 fork 只有以下 Workflow 保持 active：

- `.github/workflows/baize-ci.yml`
- `.github/workflows/baize-cache-impact.yml`
- `.github/workflows/baize-docs-impact.yml`

上游 Workflow 文件继续保留以减少合并冲突，但在 GitHub 仓库级禁用。
官方发布、签名、npm、Cloudflare、Pages、E2E、标签、定时维护和社区机器人
均不运行。GitHub 可能额外显示平台托管的
`dynamic/dependabot/update-graph`；它不是仓库自有 Workflow。禁用它会同时
关闭依赖图与漏洞告警，因此不计入 Baize allowlist 检查。

## 获取并整合上游

先只更新镜像：

```powershell
pwsh -NoProfile -File scripts/sync-upstream.ps1
```

```sh
sh scripts/sync-upstream.sh
```

脚本要求工作区干净，结束后返回原分支，验证本地 `main-v2` 与
`upstream/main-v2` 完全一致，并且绝不自动 push 或合并 `custom/baize`。
审计总体、WebUI 和 Workflow 差异后再执行：

```powershell
git push origin main-v2
git switch custom/baize
git merge --no-ff --no-commit main-v2
```

如果上游新增未知 Workflow，必须暂停。首次推送该文件前临时关闭仓库
Actions，把路径加入禁用清单并应用 GitHub 设置，之后才恢复 Actions。

Baize merge driver 会保留分叉拥有的前端文件。后端冲突正常解决；需要的
上游 WebUI 变化手动移植，并更新 `docs/UPSTREAM_SYNC_LOG.md`。未提交的失败
合并使用 `git merge --abort`；已共享合并使用 revert，不得改写历史。

## WebUI 所有权与验证

`.gitattributes` 标记 `merge=baize` 的主/登录页面、Logo、`baize.css` 和
`baize.js` 由本分叉维护。HTML、CSS、JavaScript 分别承载结构、样式和行为；
后端路由与 API 始终正常参与三方合并并配套 Go 测试。

提交使用 Conventional Commits。功能提交前运行聚焦测试、gofmt 和
`git diff --check`；上游整合前运行 `go vet ./...`、repolint、`go test ./...`，
WebUI 有变化时做浏览器验收。同步日志必须记录新旧 SHA、移植/忽略内容、
冲突和验证结果。
