# findings.md — 去构建链（本地 vendor min 版）

## 依赖库真实文件形态（来自 web/node_modules）
- `alpinejs@3.17.5/dist/module.esm.min.js` → **ESM**，default export `Alpine`，**不自动 start**（需用 `Alpine.start()`）。
  - 注意：`cdn.min.js` 会自动 start，会破坏"先注册 store/组件再 start"的时序，故必须用 module 版。
- `marked@18.1.3/lib/marked.umd.js` → UMD，挂全局 `marked`（该版本无 `.min.js`，只有 umd/esm，umd 体积可接受）。
- `dompurify@3.4.17/dist/purify.min.js` → UMD min，挂全局 `DOMPurify`。
- `microlight@0.0.7/microlight.js` → UMD（IIFE），挂全局 `window.microlight`（无 min 版，文件本身很小）。

## 加载方案（最终）
- `index.html`：
  - `<link rel="stylesheet" href="/src/styles/app.css">`（替代 Vite 的 css import）
  - `<script type="importmap">` 仅映射 `alpinejs` → `/vendor/alpine.module.esm.min.js`
  - 三个 UMD `<script defer>`：`/vendor/marked.umd.js`、`/vendor/purify.min.js`、`/vendor/microlight.js`
  - `<script type="module" src="/src/main.js">`
- `main.js`：`import Alpine from 'alpinejs'`（经 importmap 解析，无需改该行）；删除 `import './styles/app.css'`；
  去掉 8 个 `import x from './views/*.html?raw'`，改 `await loadTemplates()`（fetch `/src/views/*.html` 与 `/src/views/partials/*.html`）填充 PARTIALS/VIEWS。
- `markdown.js`：删三行 vendor import，改为 `const marked = window.marked; const DOMPurify = window.DOMPurify; const microlight = window.microlight;`（在模块顶层读取，UMD script 已先执行挂好全局）。

## 顺序保障
- 普通 `<script defer>` 与 `<script type="module">` 都在 DOMContentLoaded 前按文档顺序执行；3 个 UMD defer 先于 main.js（module）执行，
  因此 main.js 启动时 `window.marked`/`DOMPurify`/`microlight` 已就绪。alpine 通过 module import 加载，无全局副作用，start 由 main.js 显式调用。

## Go 端
- `//go:embed web`（原 `web/dist`）；`distSub()` → `fs.Sub(efs, "web")`。
- 现有 `e.GET("/*")` 已：URL path 在 embed FS 存在即 `c.FileFS(name, dist)` 返回；否则回退 `index.html`。无需新增路由即可 serve `/src/*`、`/vendor/*`、`/src/views/*.html`、`/favicon.svg`。

## 安全边界（不变）
- `x-markdown` 仍走 marked → DOMPurify.sanitize → microlight.reset()；`enhanceMarkdown` 仍先快照 textContent，异常恢复。
- `safeColor()` / `safeFrom()` / 同源 `navigate()` 保持不变。

## 删除清单（均为生成物/依赖，非源码）
`web/node_modules`、`web/dist`、`web/public`（svg 已复制）、`web/package.json`、`web/pnpm-lock.yaml`、`web/pnpm-workspace.yaml`、`web/vite.config.js`、`web/.DS_Store`。

## 验证期已知坑
- sandbox 外网不可达 → 只能从 node_modules 复制，版本锁定于 package.json。
- embed 前务必删 node_modules，否则 embed 失败/二进制暴涨。
- fetch 模板是运行时异步，main.js 用 top-level await（module 支持）。
