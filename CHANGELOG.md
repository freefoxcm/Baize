# Changelog

All notable changes to the Go line (Reasonix 1.0+) are recorded here. The legacy
`0.x` TypeScript history lives on the [`v1`](https://github.com/esengine/DeepSeek-Reasonix/tree/v1)
branch.

## Unreleased

### Added

- **Serve defaults tool approval to `auto`** (desktop parity): a freshly
  built serve controller applies `desktop.default_tool_approval_mode`
  (auto unless configured otherwise) instead of the kernel's conservative
  ask. Runtime modebar switches are unaffected.
- **Error messages width-bounded**: `.msg--error` now respects
  `--chat-maxw` like regular messages, so notices such as
  `✗ context canceled` no longer stretch across the full transcript width.
- **Ask card aligned with desktop AskCard**: questions render as a
  row-list (numbered options + "Other answer" expandable row + red
  "Skip and keep chatting" row), a confirm bar with a dynamic label
  (Next / Submit / Skip-and-keep-chatting), answered-summary crumbs,
  a Back pill in a quick-actions row, a `q.header` badge and a
  Stop-task button, a "Question 1/2" progress badge, and desktop keyboard
  parity (digits, arrows, Enter, ←/Backspace back, **Esc now stops the
  task** via `POST /cancel` instead of skipping).

- **`/effort` command fixed**: a bare `/effort` submit is now intercepted by
  the serve backend and reports the current reasoning-effort capability (same
  payload as `GET /effort`) instead of falling through to the controller's
  "unknown command" notice — symmetric with how `/model` lists models.
- **Ask decision shelf (desktop `.prompt-shelf` parity)**: questions are no
  longer stacked into the transcript. `showAsk` renders a single card into a
  new `#ask-slot` in the footer (above the composer, `--chat-maxw`-centered,
  `max-height:min(82vh,720px)` scrollable). Multi-question asks advance one
  question at a time with a `1/N` progress badge, Back, custom-answer input,
  Skip / just-chat, and full keyboard support (digits 1-9, arrows, Enter,
  Esc). Concurrent asks queue instead of racing: answers are only sent via
  `POST /answer` once all questions are answered, and the wait-timer resumes
  exactly once.

- Welcome-page **Token activity** uses the Codex-style seven-row, week-column
  heatmap with theme-derived activity levels and presets for the last six
  months (the default), the last three months, and this year. Hover, focus, or
  tap a
  day for a floating requests / turns / per-model breakdown without shifting
  the welcome layout. The new `GET /usage/calendar?range=year|6m|3m`
  endpoint aggregates existing `serve` stats rows, so historical usage needs
  no new ledger. Long Windows and POSIX workspace paths now display their final
  folder while keeping the complete path available as hover/accessibility text.
  Model listing is unified: `/model` output de-duplicates identical model
  names across provider aliases (deepseek / deepseek-flash / deepseek-pro),
  matching the frontend catalog; and typing `/model ` in the composer opens an
  argument palette (desktop ArgMenu) with the de-duplicated model refs — type
  to filter, arrows + Enter or click to fill the ref, current model tagged.
- Run-strip desktop parity: the ticker now counts NET work time (approval/ask
  waits are paused and excluded via `waitAccumMs`), waits freeze the ticker and
  show a stable status line, and retry shows only its own copy. The footer
  status text is back to pure connection state (connected/reconnecting/
  disconnected) — running state lives solely in the composer run-strip. The
  strip matches the desktop styles (accent color, 6px currentColor dot, warn
  waiting tint, `tabular-nums`) and gained `sr-only` announce + `aria-hidden`
  ticker for accessibility.
  The footer toolbar is gone: the connection dot/label, turn-info and the
  duplicate balance display were all redundant with the sidebar status block.
  Connection state now lives solely in the sidebar dot (with a tooltip), the
  approval slot sits directly above the composer card, and the duplicate
  `#approval-slot` element was removed.
  Approval-mode parity with desktop: the ask/auto/yolo modebar stays usable
  while a turn is running (only a decision surface — an approval/ask card —
  disables it), and plan approvals (`exit_plan_mode`, a fresh human decision)
  render as a dedicated card ("Approve plan" / deny, no session/persist
  grants) with a matching run-strip "Waiting for plan approval…" line.
  `TestServePlanApprovalPostureMatrix` proves the plan card surfaces over HTTP
  in ask, auto, and yolo alike.
  The slash-command palette now anchors to the composer card (previously the
  full-width footer), so it matches the input width exactly and opens just
  above it — desktop `.slashmenu` positioning (`bottom: calc(100% + 6px)`,
  `left: 0; right: 0`), with a 360px/50vh height cap.
  The sidebar nav is trimmed to "New session" only: the compact / rewind /
  branches / models entries are gone (their modals remain reachable via slash
  commands), and the stats entry moved into the status block at the bottom of
  the sidebar as a bare icon button next to the connection dot/model.
  The stats icon now sits at the right of the "Status" heading (with a hover
  tooltip), the sidebar nav no longer claims flex space (so the session list
  grows to fill the freed height), and `applyStaticI18n` now translates
  `aria-label` attributes too.

- Added a **Remote SSH** module (VS Code Remote-SSH style): a user-global
  `[remote]` host config, `reasonix remote` CLI (add/list/remove/import/test/
  connect/status/forward/serve/fs) and `/remote` slash command, an SSH transport
  with trust-on-first-use host-key verification, keepalive + exponential-backoff
  reconnect, `-L`/`-R` port forwarding, and SFTP file access. `connect`
  bootstraps a persistent `reasonix serve` on the remote host and tunnels its
  loopback port so the full agent runs remotely. The desktop app adds a
  **Settings -> Remote SSH** host manager, a remote file browser/editor, a
  port-forwarding panel, and a status-bar connection chip. Linux/macOS remotes.
- Added `reasonix serve --port-file/--token-file/--pid-file` so a supervised
  headless serve can bind an ephemeral port and read its auth token from a file
  (keeping it out of `ps`).
- The serve WebUI tool-call cards and turn summary bars now cap at the same
  width as messages (760px), so history-rebuilt cards no longer overhang the
  transcript. Tool approval requests render as a desktop-style decision shelf
  (amber tool badge, monospace subject block, single-column action list with
  `1-4` number keys, select-then-confirm bar) instead of the old horizontal
  button row, and usage events no longer emit metric rows into the chat —
  tokens/cost/cache accumulate into the footer ticker, sidebar, and stats
  modal, matching the desktop app. Approval shelves with long subjects or
  reasons now cap their body/action regions with scrollable max-heights (the
  confirm bar stays visible), and continuing an older session from the history
  rebuild correctly opens a fresh turn container for new messages — the
  rebuild also gates in-flight SSE events (`historyPending`) so nothing is
  lost or double-rendered while the transcript reloads. Approval action
  buttons now use fixed labels with a short description line (the full command
  stays in the subject block; the matched rule is shown as a tooltip), and the
  approval shelf is pinned to a fixed slot above the composer instead of
  scrolling with the transcript — a new request replaces the previous card.
  The shelf now sits directly below the to-dos panel (desktop footer order),
  and reasoning blocks match the desktop ReasoningPanel: a single head button
  ("Thinking…" with a shimmer while streaming, "Thinking done · Ns" after),
  auto-collapse that a manual toggle overrides, a left-border body without an
  inner scrollbar, streaming truncation at 12k chars / 240 lines, and no copy
  button. The chat column width is now a single `--chat-maxw` variable set to
  960px — matching the desktop app's `--maxw` — and every constrained element
  (messages, tool cards, turn summaries, notices, approvals, todos) consumes
  it instead of nine hardcoded 760px rules. Completed turns now auto-fold
  their tool calls and reasoning behind a desktop-style count summary bar
  ("Working…" while running, "Worked · 2 tools · 1 thought" after), both for
  history rebuilds and live turns at `turn_done`; a manual toggle overrides
  the auto behavior. User messages now render as right-aligned bubbles
  (desktop style) instead of left caret lines, and `/history` strips the
  system-injected compose prefixes (plan-mode marker, language directives,
  transient blocks, referenced-context preambles) plus synthetic/empty user
  turns — so history shows the user's actual text, matching the desktop app.
  The turn fold bar now sits after the user message (desktop TurnCollapse
  position) and is styled as a borderless inline button with a hover tint
  instead of a bordered card, aligned to the message column width.
  The composer is now a desktop-style card: a run strip above the input
  (spinner-word ticker, elapsed time, live token count; "waiting for
  approval/answer" and retry states), an ↑ send / ⏹ stop button pair, and a
  bottom meta bar with an ask/auto/yolo approval modebar (toolbar auto/yolo
  buttons removed) and an inline model switcher opening the models modal.
  The composer meta bar is now fully desktop-aligned: a Direct/Plan/Goal task
  mode trigger (toolbar plan/goal buttons removed), modebar thumb colors per
  mode (ask neutral, auto blue, yolo red with white active text), a model
  switcher popover (search, provider grouping with current group first,
  check mark on the active model, `label · provider` trigger), and an effort
  switcher backed by new GET/POST /effort endpoints (hidden when the active
  provider does not support effort).
  Composer width now matches the message column (960px, desktop --maxw) and
  the card uses overflow:visible so the upward menus are no longer clipped.
  The approval modebar thumb/text details are aligned item-by-item with
  desktop (14px icons, fg-dim inactive text, 620 weight, elevated ask thumb).
  Work mode (runtime profile) is now a separate control from execution mode,
  matching desktop: execution method shows Standard/Plan/Goal, and a new Work
  mode trigger switches Balanced/Lightweight/Delivery via new GET/POST
  /profile endpoints (rebuilds the controller under boot.Options.TokenMode;
  in-memory, resets to full on restart).
  The model switcher groups by the mapped provider label instead of the raw
  provider name, so deepseek / deepseek-flash / deepseek-pro collapse into a
  single provider group; same-named models inside a group are de-duplicated
  (active/default entry wins).
- The serve WebUI now renders assistant messages with GFM markdown, syntax
  highlighting (highlight.js), collapsible reasoning blocks that auto-expand
  while streaming and auto-collapse when done, per-turn grouping with a
  foldable tool-call summary bar, unified-diff previews in tool cards, and
  collapsible tool errors. Messages get hover actions (copy, inline edit via
  the new `POST /edit` endpoint), markdown images render through the new
  workspace-confined `GET /file` endpoint with a click-to-zoom lightbox, and
  pasted/dropped images upload via `POST /attach`. Rendering libraries
  (marked + DOMPurify + highlight.js) are embedded in the binary — no network
  or build step required; rebuild with `scripts/build-serve-vendor.mjs`.
- Added Claude Code-style searchable CLI pickers for models, providers, and
  sessions, with arrow, Vim, and `Ctrl+P` / `Ctrl+N` navigation.
- Added `-p` / `--print`, `text`, `json`, and `stream-json` output modes for
  one-shot use and automation.
- Added session-scoped `--allowed-tools`, repeatable `--add-dir`, Claude-compatible
  permission modes, flexible `--resume [QUERY]`, and the `--copy` resume escape
  hatch.
- Added `/status` details for the active model, effort, cache, Git state,
  background jobs, work profile, and provider balance where available.

### Changed

- Automatic Plan Mode has been retired. Plan Mode is now always entered through
  an explicit user choice, and the one-time config v5 upgrade removes legacy
  `agent.auto_plan` and `agent.auto_plan_classifier` values so upgraded users
  receive the same behavior as new users.
- `Shift+Tab` now cycles CLI safe modes from Ask to Auto to Plan, while YOLO
  remains an independent `Ctrl+Y` toggle.
- Model, provider, resume, and approval menus now use consistent row selection;
  slash completion, help, aliases, and dispatch share one command registry.
- The full-screen CLI composer now uses theme-accented borders and a slim bar
  cursor by default, grows within the available terminal height, scrolls long
  drafts independently, and preserves selections across explicit image paste.
- The persistent CLI footer now uses a responsive, theme-aware layout for
  interaction state, model, effort, localized work mode, Git identity, cache,
  context, compaction headroom, jobs, and balance. Narrow terminals move or
  compact complete groups instead of clipping labels.
- CLI clipboard actions now separate terminal-native text paste from explicit
  image paste: `Ctrl+V` on macOS/Linux, `Alt+V` on Windows, or `/paste-image`.
  Local transcript copy verifies the native clipboard write, while SSH uses a
  clearly labelled OSC 52 fallback.
- Runtime rebuilds after model, effort, or work-mode changes now preserve the
  conversation, session permission overrides, additional directories, and
  session lease ownership.
- Agent execution now monitors host-observed Todo progress automatically. A
  stalled current item receives a recovery nudge after 8 tool-call rounds with
  no new completion, unique read, command, or mutation, and pauses with saved
  work after 16. Exact repeats do not renew the progress lease; real work does.
  Two-level task lists keep the single in_progress contract: the active
  sub-step is the only current item while its phase stays pending, and the
  phase becomes in_progress to sign off only after all of its sub-steps are
  completed. A level-1 sub-step with no phase header above it is rejected.
  Executor and planner rounds now use automatic progress management. Retired
  `[agent].max_steps` and `planner_max_steps` keys remain parseable for upgrades,
  but are ignored and removed by a one-time migration so stale hidden limits
  cannot truncate new behavior. One-off CLI and unattended bot limits remain.

### Fixed

- Fixed Remote Workbench failing with only `initialize: workbench-desktop:
  connection closed` on fresh or cross-platform SSH hosts. Desktop now proves
  the exact Host CLI Build ID, provisions the matching verified release without
  requiring remote npm, runs the managed binary explicitly, and preserves a
  safe structured bootstrap error when the remote command exits early.
- Hardened Bash permission reuse for dynamic and indirect execution. Parameter/arithmetic expansions,
  assignments, redirects, heredocs, and globs can only be remembered as exact
  `Bash=<literal>` rules, while still using Auto's normal fallback. Nested or
  indirect execution now requires a human in interactive Ask/Auto and fails
  closed in headless Ask/Auto/DontAsk. Broad Bash rules, Guardian/hook allows,
  and the approved-plan window can no longer silently authorize that stricter
  class; YOLO remains the explicit full-access bypass and sandbox enforcement
  is unchanged.
- Fixed Desktop sessions incorrectly locking themselves during Goal + Delivery
  mode changes, controller rebuilds, duplicate-tab restore, and background
  reattachment. Desktop now keeps one process-local runtime owner per canonical
  session, fences stale controller events by runtime epoch, blocks sends until
  that runtime is ready, and scopes single-instance ownership to
  `REASONIX_HOME` instead of the executable path. Switching saved sessions is
  now transactional: a target build, restore, or lease failure leaves the
  current controller, lease, path, mode profile, and runtime epoch untouched.
- Stabilized the desktop rich composer caret after skill and plugin invocation
  tags. DOM→model and model→DOM selection mapping now treat invocation chips as
  zero-length atoms while still counting user text that lands inside the NBSP
  caret anchor (common on Windows WebView2), restore both selection ends, and
  recover the insertion point from a `beforeinput` snapshot when the browser
  temporarily loses selection — so mid-text edits no longer jump to the end.
- Isolated the Windows desktop WebView2 shell from stale system proxies, so an
  exited proxy client cannot leave the embedded UI hidden during startup. If
  WebView2 still does not reach DOM-ready within 15 seconds, Reasonix now shows
  the native window with a recovery prompt instead of appearing not to launch.
  Remote Markdown images are fetched by the backend with Reasonix's proxy
  configuration instead of bypassing that proxy through the isolated WebView.
- Restored captured-mouse right-click text paste, made composer drag selection
  copy through the verified native clipboard path, and kept non-Git footer
  telemetry left-aligned without reserving an empty data band.
- Restored stateful MCP behavior after the v1.17.13 regression: user-added
  servers work without extra trust settings (including delivery-mode on-demand
  calls), repository-provided servers use one exact launch confirmation, and
  stdio tools reuse one persistent process so browser sessions survive across
  calls without repeated startup latency. The former trust/reverify/catalog
  management UI and CLI are removed.
- Localized persistent-footer labels and displayed work-mode values in English,
  Simplified Chinese, and Traditional Chinese, while keeping command arguments
  stable.
- Restored the `0.53` content boundary: model output, tool output, session
  transcripts, recovery branches, and background-job artifacts retain their
  original text instead of being rewritten by heuristic secret redaction.
  Credential masking remains in key-entry summaries and explicit diagnostic or
  session-cleanup paths. Transcript-bearing session/job sidecars are kept
  private (`0600`, with private job directories), and the retired
  `redact_tool_output` setting is removed with a one-time upgrade notice.

## [1.0.0] — 2026-06-03

First stable release — a **ground-up rewrite in Go**. Not an upgrade of the `0.x`
TypeScript line; a new codebase that becomes the default (`main-v2`).

### Highlights

- **Go kernel**: a single static binary (CGO-free), cross-compiled for
  darwin/linux/windows on amd64 + arm64. Distributed via npm (the package wraps
  the native binary), Homebrew (`esengine/reasonix` tap), and release archives;
  no Node runtime needed to run it.
- **Agent core**: the loop, built-in tools (read/write/edit/multi_edit/glob/grep/
  ls/bash/web_fetch/todo_write), permission gate, sandboxed bash, and the
  DeepSeek prefix-cache–oriented design.
- **Subagents**: `task` plus explore/research/review/security_review skill agents.
- **Skills & hooks**: Claude-Code-style skills (`internal/skill`) and hooks
  (`internal/hook`), symlink-aware and slash-integrated.
- **MCP client**: connect external servers over stdio / Streamable HTTP; reads
  `[[plugins]]` and a Claude-Code `.mcp.json`.
- **Code intelligence via CodeGraph**: a tree-sitter symbol/call graph
  (`codegraph_*` tools) replaces embedding semantic search — no embedding service
  or API cost. Fetched into a local cache on first use (or `reasonix codegraph
  install`) and indexed in the background, so installs and startup stay fast.
- **Plan mode** with evidence-backed step sign-off (`complete_step`).
- **Memory**: `REASONIX.md` hierarchy + auto-memory, folded into the cache-stable
  prefix.
- **ACP** (`reasonix acp`) and an HTTP/SSE server frontend; desktop app (Wails).

### Fixed

- **File encoding support restored** — GBK/GB18030 (and other non-UTF-8) files
  can now be read, edited, and grepped correctly. The v2 rewrite had dropped
  v1's encoding detection; files in CJK Windows charsets were silently misread
  or rejected as binary. The read/edit/write round-trip now preserves the
  original file encoding. (#2637)

### Notes

- Versions: the legacy TypeScript line stays in `0.x`; the Go line starts at
  `1.0.0`. See [docs/MIGRATING.md](docs/MIGRATING.md).
- Release archives ship a bare binary; CodeGraph is fetched on first use. Windows
  support for the fetched runtime is unverified — install `codegraph` on PATH if
  the auto-fetch doesn't resolve there.

[1.0.0]: https://github.com/esengine/DeepSeek-Reasonix/releases/tag/v1.0.0
