# task_plan.md — 前端去构建链：本地 vendor min 版（无 pnpm / Vite）

## 目标
1. 彻底移除 pnpm + Vite，前端改为原生 ESM + 本地 vendor 库（min 版），由 Go `//go:embed web` 直接嵌入源码目录，无需任何构建步骤。
2. 依赖库（alpine 3.17.5 / marked 18.1.3 / dompurify 3.4.17 / microlight 0.0.7）min 文件落到 `web/vendor/`，随二进制 embed 部署。
3. 所有页面、交互、API 契约、安全边界（DOMPurify 消毒、microlight 高亮、safeColor/safeFrom）保持不变。

## 阶段
- [x] S1 调研：读 main.go / main.js / markdown.js / index.html / Makefile / package.json；确认依赖文件形态；外网不可达改用 node_modules 复制。
- [x] S2 准备：复制 `web/vendor/*`（4 库）+ `web/favicon.svg` / `web/logo-forest.svg`。
- [x] S3 改造：`main.go`（embed→web）、`main.js`（fetch 模板 + 去 css import）、`markdown.js`（去 vendor import 改全局）、`index.html`（importmap + UMD script + link css）、`Makefile`（去 pnpm/vite）。
- [x] S4 清理：删 `web/node_modules`、`web/dist`、`web/public`、`package.json`、`pnpm-lock.yaml`、`pnpm-workspace.yaml`、`vite.config.js`、`.DS_Store`。
- [x] S5 验证：`go test ./...` + `go build`（embed web）+ HTTP 冒烟（全 200）+ vendor 文件格式/符号校验。
- [x] S6 收尾：更新 README / ADR 0004 / planning / workspace memory；停验证服务。

## 约束 / 风险（已处理）
- sandbox 外网不可达（curl 000）→ 从 `web/node_modules` 复制锁定版本；embed 前必删 node_modules。
- Alpine 必须用 ESM module 版才能手动 `Alpine.start()`（UMD cdn 会自动 start）；marked/dompurify/microlight 用 UMD 全局。
- 模板 `?raw` 改为运行时 `fetch('/src/views/*.html')`，Go embed web 后这些文件可被静态 serve。
- 未 commit；不 `git add -A`。

## 验证结果
- `go build -o /tmp/flock-cdn .` 成功（25M，embed web）。
- `GOFLAGS=-mod=mod go test ./...` 全绿；go.mod/go.sum 无残留改动。
- HTTP 冒烟：health 200；/、/src/main.js、/vendor/*、/src/views/*.html、/src/styles/app.css、/favicon.svg、/?sort=hot、/topics/1 全部 200。
- vendor 校验：alpine `export{...as default}`、marked UMD default object、dompurify/microlight UMD 全局，版本正确。
- 沙箱无 chromium，未跑无头浏览器回归；逻辑与加载顺序（UMD defer 先于 module）经代码审查确认。
