# Baize 分叉维护规范

本仓库把“获取上游代码”和“开发 Baize 产品”分开。以下规则同时约束开发者和编码代理。

## 分支职责

| 分支 | 职责 | 允许的改动 |
|---|---|---|
| `upstream-sync/main-v2` | `upstream/main-v2` 的干净镜像 | 只能快进更新 |
| `custom/baize` | Baize 集成与部署 | 已审查的功能提交和受控上游合并 |
| `feat/*`、`fix/*` | 短期开发 | 基于 `custom/baize` 的单一完整改动 |
| `main-v2` | 保留旧历史 | 不再进行新的 Baize 开发 |

禁止在同步分支提交、禁止 rebase `custom/baize`、禁止向 `upstream` 推送，也禁止把功能开发混进上游 merge commit。

## 配置新克隆

在仓库根目录运行对应脚本一次：

```powershell
pwsh -NoProfile -File scripts/setup-fork-git.ps1
```

```sh
sh scripts/setup-fork-git.sh
```

脚本会幂等启用 `rerere`、仅快进 pull、远程清理、安装到 `.git/baize-hooks` 的跨分支持久 hooks、Baize merge driver，并把 `upstream` 设为只获取不可推送；同步分支切换后保护仍然有效。

## 获取并整合上游

先只更新干净镜像：

```powershell
pwsh -NoProfile -File scripts/sync-upstream.ps1
```

```sh
sh scripts/sync-upstream.sh
```

同步脚本要求工作区干净，结束后返回原分支，并且绝不会自动 merge、commit 或 push。先审查输出的提交范围，尤其是上游前端变化：

```powershell
$oldBase = git merge-base custom/baize upstream-sync/main-v2
git diff "$oldBase..upstream-sync/main-v2" -- internal/serve/index.html internal/serve/login.html internal/serve/serve.go
```

再开始受控合并：

```powershell
git switch custom/baize
git merge --no-ff --no-commit upstream-sync/main-v2
```

已配置的 `baize` driver 会保留本分叉拥有的前端文件。后端冲突按普通方式解决；需要的上游 WebUI 变化必须手动移植到 Baize 资源，并更新 `docs/UPSTREAM_SYNC_LOG.md`。完成全部检查后才能提交 merge，并把两个分支推送到 `origin`。

如果合并尚未准备好，在提交前使用 `git merge --abort`。已经共享的 merge 应通过普通 revert 提交回退，不得重写 `custom/baize` 历史。

## WebUI 所有权

`.gitattributes` 中标记 `merge=baize` 的主页面、登录页、Baize Logo、`assets/baize.css` 和 `assets/baize.js` 由本分叉维护。HTML 只保留结构，样式放 CSS，浏览器行为放 JavaScript；只有防止首次绘制闪烁的主题启动代码可以内联。

后端路由和 API 不受分叉合并策略保护。新的 Serve 行为应放在职责明确的 Go 文件并配套测试，让上游代码正常参与合并；不得用 Baize merge driver 隐藏后端冲突。

## 提交与验证规则

- 使用 Conventional Commits，功能、修复、WebUI、上游整合和文档保持独立可审查。
- 功能提交前运行聚焦测试、`gofmt` 和 `git diff --check`。
- 上游 merge 或 push 前运行 `go vet ./...`、`make lint`、`go test ./...`；WebUI 有变化时还要做浏览器验收。
- 同步日志必须记录新旧上游 SHA、已移植/忽略的 UI 变化和验证结果；没有日志的上游合并不算完成。
