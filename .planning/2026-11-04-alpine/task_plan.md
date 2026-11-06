# task_plan.md — 前端改写为 AlpineJS 3 + 清理旧代码

## 目标
1. 将 `web/` 前端从 React 18 + Chakra UI + Vite 完整改写为 Alpine.js **3.17.5**（官方 registry 可安装的 latest）+ Vite 8 纯 JS 构建。
   - 初始 WebSearch 摘要称 `3.17.10`，经 `registry.npmjs.org/alpinejs/latest`、dist-tags、镜像源与实际 `corepack pnpm install` 四方核实，实际 latest 为 **3.17.5**。差异记入 findings.md 与 ADR 0004。
2. 保持现有全部页面、交互与后端 API 契约不变，Go `embed.FS` 仍从 `web/dist` 取产物。
3. 改写完成后删除旧 React 代码、依赖与配置，更新 README / Makefile / ADR。
4. 用户追加要求：代码高亮改用 npm 最新版 `microlight 0.0.7`，替换初版 `highlight.js`。
5. 跑通质量门：`pnpm run build`、`go test ./...`、启动后 HTTP 冒烟 + 无头浏览器逐页/交互回归。

## 阶段
- [x] S1 勘察：读完 `web/src/` 全部源码 + handlers DTO + 视觉参考；核对 npm 版本与工具链。
- [x] S2 设计：确定目录/路由/状态/Markdown/通知/样式方案，记入 findings.md + ADR 0004。
- [x] S3 实现：新 `web/index.html` + `app.css` + JS（core: api/store/router/markdown/utils；pages/components/views）。
- [x] S4 构建与验证：依赖安装、构建、Go 测试、启动服务 + HTTP 冒烟 + Puppeteer 逐页/交互回归。
- [x] S5 清理：删旧 tsx/js/依赖/锁文件，更新 package.json、vite.config.js、Makefile、README、main.go。
- [x] S6 microlight 替换：核实 `microlight@0.0.7` 为 latest，替换 `highlight.js`，更新 markdown.js / README / ADR 0004，重建产物。
- [x] S7 收尾：修复分页 getter spread、编辑器光标两个验证期 bug；复跑全部质量门；写 progress/findings/workspace memory；停掉验证服务。

## 约束 / 已知风险（均已处理）
- `git status` 已有他人暂存的改动：`Makefile`、`web/package.json`、`web/pnpm-lock.yaml`、`web/pnpm-workspace.yaml`。
  只改本次涉及的文件；未 `git add -A` / `reset --hard` / `clean -fd`；未主动 commit。
- `web/dist` 被 `main.go` 的 `//go:embed web/dist` 引用 → `emptyOutDir: true` 重建后目录内始终有 `index.html`。
- Markdown 内容来自用户，是可信边界：`x-markdown` 先 `marked` → `DOMPurify.sanitize` → `microlight.reset()`。
- 表单校验、鉴权 token 处理、错误提示、a11y（label/aria）均保留；新增 `safeColor()` / `safeFrom()` / 同源 `navigate()` 校验。
- `go test ./...` 需要 `go mod tidy`（依赖版本漂移），用 `GOFLAGS=-mod=mod` + go.mod/go.sum 备份与 trap 恢复，最终两文件无改动。

## 验证命令（实际执行形式）
- `cd /opt/repos/flock/web && corepack pnpm install && corepack pnpm run build`
- `cd /opt/repos/flock && GOFLAGS=-mod=mod /usr/local/go/bin/go test ./...`
- `GOFLAGS=-mod=mod /usr/local/go/bin/go build -o /tmp/flock-final2 .` → 隔离目录 + 8091 端口启动
- `curl http://localhost:8091/api/v1/health`、首页 / SPA fallback / assets
- Puppeteer：8 个页面渲染、登录/排序/分类/回复/弹窗/编辑器/标签/发布/后台 5 tab/通知下拉、XSS 与 microlight 专项

## 状态
全部阶段完成。详见 `progress.md` 的最终验证记录与 `findings.md` 的结论。
