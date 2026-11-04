# ADR 0002: 前端拆包 + 改用 md-editor/nohighlight

## Status
Superseded by ADR 0004（前端已迁至 Alpine.js，React 生态拆包策略整体作废）

## Context
- 打包产物是单个 `index-*.js`，minify 后 1.6MB / gzip 562K，加载负担过大。
- 目标是拆成多个 chunk，且每个 gzip 后尽量 ≤100K。
- `react-markdown` + `@uiw/react-md-editor` 默认版本会经 `rehype-prism-plus` + `refractor` 打包全部语言语法（约 230K gzip），是最大的单块。

## Decision
- 在 `web/vite.config.js` 用 `manualChunks` 按依赖库体系拆包：
  react / router / chakra / emotion / motion / static(state,zustand) / icons /
  react-md / remark / rehype / rehype-parse(parse5) / rehype-sanitize / chain-jsx /
  micromark / md-base / md-editor / utils / index(app)。
- 三处导入改为 `@uiw/react-md-editor/nohighlight`，从根源移除 CodeMirror 与 prism 高亮。
- `vite dev` 代理目标改为读仓库根目录 `.env` 的 `SERVER_PORT`/`BASE_URL`，使 `make dev` 的后端端口与前端代理一致（此前硬编码 8080，与 .env 的 8090 不符）。

## Consequences
- 正：全部 JS chunk gzip ≤100K（最大 react 85.7K）。
- 负（ceiling）：Markdown 编辑器的**预览/编辑区失去语法高亮着色**（textarea 模式）。
  如需恢复高亮，可单独懒加载 `rehype-prism-plus` 并按需注册少数语言，不应再全量引入 refractor。
- 负：存在少量 harmless 的 rollup "Circular chunk" 提示，不影响产物。
- `make dev` 现与 `make run`/生产一致地使用 `.env`（后端本就读 .env，此处补齐前端代理）。