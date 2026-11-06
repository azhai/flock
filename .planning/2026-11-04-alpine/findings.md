# findings.md — AlpineJS 改写

## 版本结论（已核实）
- WebSearch 摘要声称 alpinejs `3.17.10`（"2 days ago"），但**官方 registry 与可安装源不一致**：
  - `https://registry.npmjs.org/alpinejs/latest` → **3.17.5**
  - `registry.npmjs.org/alpinejs` dist-tags → `{"latest":"3.17.5"}`，版本列表止于 3.17.5
  - 配置的镜像 `https://registry.npmmirror.com/` → `npm view alpinejs version` 也是 3.17.5
  - `corepack pnpm install` 实际装到 3.17.5
  - 结论：**以可实际安装的最新版 3.17.5 为准**（已写入 package.json 精确版本、ADR 0004）。
- `microlight` npm latest = **0.0.7**（UMD 包，导出 `reset()`，无语言包）。按用户要求替换 `highlight.js 11.12.0`。
- 最终依赖：alpinejs 3.17.5、marked 18.1.3、dompurify 3.4.17、microlight 0.0.7、vite 8.3.5（dev）。共 4 运行时 + 1 构建依赖（原 17 个）。
- Node v22.22.2，`pnpm` 不在 PATH，用 `corepack pnpm`（packageManager `pnpm@12.11.2` 为他人暂存改动，保留）。
- Go 1.27.3 在 `/usr/local/go/bin/go`，不在 PATH；`go test` 需要 `GOFLAGS=-mod=mod`。
- go.mod 的本地 replace `../gobus`、`../goent` 均存在。

## 结构 / 契约
- `main.go` `//go:embed web/dist`，`e.GET("/*")` 非 `/api/` 路径回退 `index.html` → History API SPA 路由可行，产物必须是 `web/dist/index.html`。
- API 前缀 `/api/v1`，`helpers.NewPaginatedResponse` → `{data,total,page,per_page,...}`；错误体 `{error: "..."}`。
- 登录态：localStorage `token` + `user`，请求头 `Authorization: Bearer <token>`。
- 通知 SSE：`GET /api/v1/notifications/stream` 需 Authorization 头，用 **fetch 流**（EventSource 无法带自定义头），断线回退 30s 轮询。
- 路由表：`/`、`/topics/:id`、`/new`、`/login`、`/register`、`/u/:username`、`/bookmarks`、`/admin`，其他 → `/`。
- 首页 query：`q`(FTS5)、`cat`(slug→category_id)、`tag`、`sort`(latest|hot|top|likes)、`page`；话题页 `back` 参数用于返回时还原筛选。
- 分类/标签颜色来自后端并拼进 style → `safeColor()` 只接受 `#rgb` / `#rrggbb` / `#rrggbbaa`（信任边界）。

## 旧前端功能清单（逐页）
- Header：Logo + 滚动后的话题标题、搜索（回车跳 `/?q=`）、新话题、通知下拉（未读角标/全部已读/点击跳话题）、用户菜单（主页/收藏/后台/退出）。
- Sidebar：分类（选中态 clip-path 箭头）+ 热门标签（前 12 个，点击切 `?cat=`/`?tag=`）。
- Home：排序 tab、话题列表（置顶/分类徽章/参与者头像/标签/点赞/浏览/时间）、空态、骨架屏、高级分页（首末页/上一页下一页/页号窗口/跳页）。
- TopicDetail：主帖 + 回复流、内部滚动容器 + 右侧「钓鱼线」刻度条（拖拽/滚轮/点击跳帖/吸附）、Header 标题随滚动接管、右侧元信息栏（状态/分类/标签/作者等级/收藏关注置顶锁定隐藏更改/返回）、回复框、帖子编辑/删除、顶踩投票、举报、编辑话题（标题/分类/标签）、锁定话题禁止回复、两个 Modal。
- NewTopic：标题/分类/标签输入（回车逗号空格提交、建议、最多 8 个、Backspace 删除）、Markdown 编辑器（工具栏/表情/预览/字数）、未登录引导。
- Login/Register：表单校验、toast、登录后回来源页（`safeFrom()` 防开放重定向）。
- Bookmarks / UserProfile：收藏列表 / 用户信息 + TA 的话题。
- Admin：5 tab（概览统计、分类 CRUD、举报流（含帖子 Markdown 预览/处理/驳回）、话题管理（隐藏/置顶/锁定/删除/分页）、用户管理（角色/启停/搜索/分页））。
- Markdown：GFM + 单回车换行（breaks）+ 代码块高亮；**内容来自用户，必须消毒**。

## 设计决定
- 纯 Vite + ESM，无 React 插件；页面模板用 `.html?raw` 导入，启动时注入 `#views`，各页用 `<template x-if="$store.route.name==='x'">` 切换 → 等价于组件挂载/卸载，切换时 `init()` 重新拉数据。
- 核心 seam：`core/api.js`（HTTP）、`core/router.js`（路由）、`core/stores.js`（session/ui/route/notifications）、`core/markdown.js`（`x-markdown` 指令：marked → DOMPurify → microlight）、`core/utils.js`（格式化/分页/徽章/编辑器 mixin）。
- 公共片段（sidebar/pagination/editor/tag-input）用 `<!-- @include xxx {"field":"..."} -->` 占位替换，避免重复模板。
- 图标用内联 SVG sprite + 少量文字符号，去掉 react-icons。
- 样式：手写一份 `app.css`，复用旧 `markdown.css` 的配色与排版规则；microlight 直接给 `span` 写 inline style，无需 `.hljs-*` / `.token-*` class 规则。
- 保留 `pnpm-workspace.yaml`（`allowBuilds: esbuild`）与 Makefile 的 pnpm 命令，属他人改动。
- 避开 `x-modelable` 等跨组件绑定：编辑器状态放宿主页面 data，`postRow` 用共享对象 + 事件通信。

## microlight 实现要点
- `microlight.reset()` 全局扫描 `.microlight` 元素并基于 `textContent` 重绘 inline-style `span`，语言无关。
- `enhanceMarkdown()` 给 `pre code` 加 `.microlight`，先快照 `textContent`，调用 `reset()`；异常时按快照恢复纯文本（降级不丢内容）。
- 因 `reset()` 是全局扫描，每次预览/帖子渲染都会重扫一遍已着色的 code，故用 textContent 快照做幂等保护。
- 外链统一 `target="_blank" rel="noopener noreferrer"`。
- 体积收益：JS 从 `349.05 kB / gzip 114.43 kB`（highlight.js 版）降到 `196.68 kB / gzip 63.40 kB`。

## 验证期发现并修复的问题
- `paginationMixin` 用 getter `totalPages`，被 `...paginationMixin` 展开后 getter 退化成静态值 → 分页显示 `NaN`。改为方法 `totalPages()` 并同步改 `pagination.html`。
- `editorMixin.toolbar()` 无选中文本时光标落在 `****` 末尾 → 改为 `setRangeText(..., 'end')` + `setSelectionRange(start + before.length, ...)`，输入落在 `**X**` 中间。
- `topic.html` 的钓鱼线 rail 初版放在 `.post-scroll` 内会随内容滚动 → 移到滚动容器外、仍在 `.topic-main-col` 内。
- `navigate()` 对外部 origin 会 `location.href` 跳转、路由 params 未解码、`from` 可形成开放重定向 → 同源校验 + `decodeURIComponent` + `safeFrom()`。
- 8090 被既有 flock 占用、IPv4 8091 被 `bakery.da` 占用 → 用 8091 IPv6 / `localhost` 隔离验证，未杀他人进程。
- 重复点击收藏/投票是幂等切换，第二次点击为取消，非 bug。

## 最终验证结论
- `corepack pnpm run build`：`✓ 34 modules transformed`，`index-DPVHNkGW.js` 196.68 kB / gzip 63.40 kB，`index-D1wKo7Ev.css` 23.47 kB / gzip 5.51 kB。
- `GOFLAGS=-mod=mod go test ./...`：6 个包全绿；`go build` 成功；go.mod / go.sum 无残留改动。
- HTTP 冒烟：`/api/v1/health` → `{"status":"ok"}`，首页引用新 assets，SPA fallback 与 JS 均 200。
- Puppeteer：home/topic/login/register/new/admin/bookmarks/profile/search 全部 `renderedViews:1`，`ERRORS []`；交互（登录、排序、分类、回复、弹窗、编辑器、标签、发布、后台 5 tab、通知下拉、SPA 无整页刷新）通过。
- XSS 专项：`{"executed":false,"scriptTags":0,"jsHrefs":0,"headings":1,"microSpans":7,"microClass":1}`；代码块文本完整。
- 分页/光标专项：`共 100 条 · 第 1 / 5 页`、跳页到第 2 页、`abc**X**`、microlight 1 个代码块 / 7 个 span / 文本保留。
- 截图：`/tmp/flock-browser/home.png`、`topic.png`、`microlight.png` 无明显错位。

## 遗留说明
- `.trae/documents`、`.codebuddy/memory` 为历史文档，保留其中的 React 语境不动；当前生效的 ADR 0001/0002 已标注前端技术栈被 ADR 0004 取代。
- `web/src`、`web/package.json` 无 React 依赖或引用；未 commit（遵循 git 安全规则）。
