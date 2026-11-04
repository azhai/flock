# ADR 0004: 前端从 React 迁移到 Alpine.js 3

## Status
Accepted（2026-11-04）｜Supersedes ADR 0002 的拆包策略部分

## Context
`web/` 原为 React 18 + Chakra UI + Zustand + react-router + @uiw/react-md-editor + react-markdown，
约 3500 行 TSX，构建链（Vite 5 + @vitejs/plugin-react + 复杂 `manualChunks`）主要是在为
React/Markdown 生态的体积买单。需求是把前端改用最新版 Alpine.js 并清理旧代码，
同时保持 Go `embed.FS` 单二进制部署与全部 API 契约。

版本核查：WebSearch 的 npmjs 摘要显示 `3.17.10`，但官方
`https://registry.npmjs.org/alpinejs/latest` 与实际安装源（npmmirror）都是 **3.17.5**，
版本列表也止于 3.17.5。以**可实际安装**的 `latest` = 3.17.5 为准，写入 `package.json` 精确版本。

## Decision
1. 使用 **alpinejs 3.17.5**（npm ESM，`Alpine.start()`），Vite 8 纯 JS 构建，产物仍是 `web/dist/index.html`。
2. 移除全部 React 生态依赖；Markdown 改为 `marked`(GFM + breaks) → `DOMPurify`(消毒) →
   `microlight 0.0.7`(高亮)，封装为 `x-markdown` 指令；编辑器改为原生 textarea + 工具栏。
   （高亮最初用 highlight.js，按用户要求改为 npm 上最新的 microlight 0.0.7。）
3. 路由为自建 History API seam（正则路由表 + 站内 `<a>` 点击委托），状态用 `Alpine.store`：
   `session` / `ui` / `route` / `notifications`。
4. 页面模板为 `.html?raw` 片段，启动时注入 `#views`，`<template x-if>` 控制挂载；
   公共片段用 `<!-- @include -->` 占位替换。组件逻辑用 `Alpine.data` 注册。
5. 视觉用手写 CSS 复刻（accent `#0d8a6f`、水彩渐变背景），复用旧 markdown.css 的排版与配色规则。

## Consequences
- 更易：依赖从 17 个降到 5 个，产物体积与构建复杂度大幅下降；无 CSS-in-JS / 无虚拟 DOM；
  模板即 HTML，改样式不用碰 JS；ADR 0002 的 `manualChunks` / refractor TDZ 规避全部作废。
- 更难：路由、模态框、下拉菜单、编辑器工具栏需要自己实现；没有 TS 类型检查；
  `x-html` 类注入面变多，因此所有用户内容统一走 `x-markdown`（消毒）或 `x-text`（转义）。
- 接受的 ceiling：microlight 是语言无关的通用 token 着色（text-shadow 发光效果），
  不做语言语法解析、不按语言包着色；编辑器预览为自绘，
  不再有 @uiw/react-md-editor 的完整工具栏（保留常用格式、表情、预览）。
- 安全不变：来自后端的颜色值拼接进 style 前做 hex 校验；Markdown 输出消毒；外链加 `noopener`。
