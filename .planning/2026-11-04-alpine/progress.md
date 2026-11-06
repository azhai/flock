# progress.md — AlpineJS 改写

| 时间 | 动作 | 结果 / 退出码 |
| --- | --- | --- |
| 2026-11-04 21:57 | 读 AGENTS.md、web 目录、package.json、vite.config.js、Makefile、main.go、routes、README、.ai-memory 计划 | 成功；确认 React/Vite + Go embed 结构 |
| 2026-11-04 21:57 | WebSearch Alpine.js 最新版 | 摘要称 3.17.10（**后证实有误**） |
| 2026-11-04 21:58 | 创建 task_plan.md / findings.md / progress.md | 成功 |
| 2026-11-04 22:00 | `npm view alpinejs version` + WebFetch `registry.npmjs.org/alpinejs/latest` + dist-tags | 全部为 **3.17.5**；修正版本结论写入 findings / ADR 0004 / package.json |
| 2026-11-04 22:02 | `pnpm -v` | 失败：`command not found: pnpm` → 改用 `corepack pnpm` |
| 2026-11-04 22:03 | 写 `web/index.html`、`src/main.js`、`core/{api,router,stores,markdown,utils}.js`、`components.js`、`pages/*.js`、`views/*.html`、`styles/app.css` | 成功；8 页 + 4 个 partial |
| 2026-11-04 22:10 | `cd web && corepack pnpm install && corepack pnpm run build` | 成功；装到 alpinejs 3.17.5 |
| 2026-11-04 22:12 | `go test ./...` | 失败：`command not found: go` → 用 `/usr/local/go/bin/go`（1.27.3 darwin/arm64） |
| 2026-11-04 22:12 | `GOFLAGS=-mod=mod /usr/local/go/bin/go test ./...` | 失败：`go: updates to go.mod needed; to update it: go mod tidy` → 用 `GOFLAGS=-mod=mod` + 备份 go.mod/go.sum + trap 恢复 |
| 2026-11-04 22:14 | 首次组装验证服务 `:8090` | 失败：`listen tcp :8090: bind: address already in use`（已有 flock 进程）；不杀，改用隔离目录 `/tmp/flock-verify` + 8091 |
| 2026-11-04 22:15 | `curl http://127.0.0.1:8091/` | 命中「花溪烘焙 - 管理后台」；`bakery.da` 占 IPv4 8091，新服务占 IPv6 → 改用 `localhost:8091` / `[::1]:8091` |
| 2026-11-04 22:16 | 对 8090 做 HTTP 冒烟 | 看到旧 `index-C4O-91Og.js` / `react-core-*.js`；原因是访问了既有旧服务而非新构建 → 只以 8091 新服务为准 |
| 2026-11-04 22:18 | Puppeteer 逐页渲染 | home/topic/login/register/new/admin/bookmarks/profile/search 全部 `renderedViews:1`，`ERRORS []` |
| 2026-11-04 22:20 | `topic.html` 钓鱼线 rail 位置 | 初版在 `.post-scroll` 内会随内容滚动 → 移到滚动容器外、仍在 `.topic-main-col` 内 |
| 2026-11-04 22:22 | 安全审查 | `router.js` params 加 `decodeURIComponent`；`navigate()` 非同源 return；`auth.js` 加 `safeFrom()` 防开放重定向 |
| 2026-11-04 22:25 | 用户反馈：代码高亮换成 microlight 最新版 | 查 npm registry + UMD 源码 + `npm pack microlight@0.0.7` → latest = **0.0.7** |
| 2026-11-04 22:27 | `core/pnpm install`（移除 highlight.js，加入 microlight） | 成功：`- highlight.js 11.12.0 / + microlight 0.0.7` |
| 2026-11-04 22:28 | 改 `core/markdown.js` 为 `microlight.reset()`；更新 README、ADR 0004 | 成功；删除 `ALIAS` 与 `.hljs-*` 逻辑，加 textContent 快照与异常降级 |
| 2026-11-04 22:30 | `corepack pnpm run build` | `✓ 34 modules transformed`；JS `349.05 kB/gzip 114.43 kB` → `196.68 kB/gzip 63.40 kB` |
| 2026-11-04 22:32 | 交互脚本 `interact.mjs` | 收藏/投票显示 `accent:false`、`score:0` → 确认为前一轮已点过的幂等切换，非 bug |
| 2026-11-04 22:33 | 编辑器工具栏验证 | 点粗体得到 `abc****` 且光标在末尾 → 修 `editorMixin.toolbar()`：`setRangeText(...,'end')` + `setSelectionRange` |
| 2026-11-04 22:34 | 分页验证 `verify2.mjs` | `PAGINATION {"pageNums":2,"summary":"共 100 条 · 第 1 / NaN 页"}` → `paginationMixin` getter 被 spread 退化；改为方法 `totalPages()` 并同步改 `pagination.html` |
| 2026-11-04 22:36 | 最终 `corepack pnpm run build` | `✓ 34 modules transformed`；`index-DPVHNkGW.js` 196.68 kB/gzip 63.40 kB；`index-D1wKo7Ev.css` 23.47 kB/gzip 5.51 kB |
| 2026-11-04 22:37 | `GOFLAGS=-mod=mod /usr/local/go/bin/go test ./...` | 全绿：flock 18.063s、config/handlers/helpers/middleware/models cached；`go build -o /tmp/flock-final2 .` 成功（约 25M） |
| 2026-11-04 22:38 | 8091 新服务 HTTP 冒烟 | `health {"status":"ok"}`；首页只引用新 assets；JS 与 SPA fallback 均 200 |
| 2026-11-04 22:38 | XSS + microlight 专项 | `{"executed":false,"scriptTags":0,"jsHrefs":0,"headings":1,"microSpans":7,"microClass":1,"codeText":"fmt.Println(\"hi\")\n"}` |
| 2026-11-04 22:39 | 复跑 `verify2.mjs` / `check.mjs` | `SORT /?sort=hot` 活跃；`共 100 条 · 第 1 / 5 页`；`PAGE2 {"page":2}`；`CURSOR abc**X**`；`MICROLIGHT codeBlocks:1 classed:1 spans:7 textKept:true`；`ERRORS []` |
| 2026-11-04 22:39 | 截图检查 | `/tmp/flock-browser/home.png`、`topic.png`、`microlight.png` 无明显错位；microlight 着色克制、文本完整 |
| 2026-11-04 22:39 | 文档一致性 | 更新 README、Makefile 注释、main.go 演示数据、ADR 0001/0002 Status、新增 ADR 0004；grep 确认 `web/src` 与 `web/package.json` 无 React 引用 |
| 2026-11-04 22:40 | 收尾 | 停掉 8091 上的验证服务（PID 95314），保留 `bakery.da`；补齐 task_plan / findings / progress；写 workspace memory |

## 最终质量门（复跑于收尾阶段）
- `corepack pnpm run build`：通过，产物 hash 与上一轮一致（只改文档，未改代码）。
- `GOFLAGS=-mod=mod /usr/local/go/bin/go test ./...`：通过；go.mod / go.sum 备份恢复后 `git status --short go.mod go.sum` 为空。
- `go build`：通过。
- 未 commit；未 `git add -A`；他人暂存的 Makefile / package.json / pnpm-lock.yaml / pnpm-workspace.yaml 保留。

## 错误清单（避免重复踩坑）
1. WebSearch 版本号不可作为安装依据 → 以官方 registry + 实际安装为准。
2. `pnpm` / `go` 不在 PATH → `corepack pnpm` / `/usr/local/go/bin/go`。
3. `go test` 要求 `go mod tidy` → `GOFLAGS=-mod=mod` 且必须恢复 go.mod/go.sum。
4. 8090 被占、IPv4 8091 被 `bakery.da` 占 → `localhost:8091`（IPv6）隔离。
5. Alpine mixin 被 spread 时 getter 失效 → 一律用方法。
6. `microlight.reset()` 是全局扫描且会重绘 DOM → 先快照 textContent，异常时还原。
