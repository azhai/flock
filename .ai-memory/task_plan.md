# 任务计划：Flock 前端从 React 迁移到 Mithril + Construct UI

## 目标
将 `web/` 下基于 React 18 + Chakra UI 的前端整体重写为 Mithril.js（使用 construct-ui 组件库），
保持现有视觉风格（浅色水粉画主题）与全部功能（话题/帖子/标签/分类/收藏/关注/通知/管理后台）。

## 现状（已读全部源码）
- 入口 `main.tsx` + `App.tsx`（ChakraProvider + BrowserRouter）。
- `lib/api.ts`（纯 fetch，可复用）；`lib/store.ts`（zustand）；`lib/useNotifications.ts`（SSE+轮询）；
  `lib/highlightProvider.js`（refractor 懒加载）。
- 组件：`Header`、`Sidebar`、`Pagination`、`CodeBlock`、`tag-input`。
- 页面：`Home`、`TopicDetail`(40KB，含自定义钓鱼线滚动条+吸附)、`NewTopic`、`Login`、`Register`、
  `UserProfile`、`Bookmarks`、`AdminPanel`。

## 关键决策
1. 用 `m()` hyperscript 写法（非 JSX），避免 babel 配置。
2. Markdown 渲染：`react-markdown`+`remark-*` → `marked`（breaks+gfm）+ `highlight.js`。
3. Markdown 编辑器：`@uiw/react-md-editor` → 原生 `<textarea>` + 实时预览面板。
4. 通知：单例模块（SSE + 30s 轮询），避免每次路由重连。
5. 删除 React 源码，重写为 .ts 模块；`construct-ui` 用于 Dialog/Tabs/Toast/Button/Input/Card 等。
6. TopicDetail 的钓鱼线自定义滚动条+吸附物理改为普通滚动容器 + 帖子节点导航点（简化但保留形态）。
7. 构建产物仍为 `web/dist`，保证 Go 侧 `//go:embed web/dist` 链路不变。

## 步骤
- [ ] 1. 规划文件 + 抓取 construct-ui API
- [ ] 2. 脚手架：package.json / tsconfig / vite.config / index.html / CSS
- [ ] 3. 核心 lib：api.ts / store.ts / notify.ts / md.ts
- [ ] 4. 组件：Layout / Header / Sidebar / Pagination / MarkdownView / TagInput / Modal
- [ ] 5. 页面：Home / TopicDetail / NewTopic / Login / Register / UserProfile / Bookmarks / AdminPanel
- [ ] 6. 入口 main.ts（m.route）
- [ ] 7. 删除旧 React 文件；`npm install`；`npm run build` 验证 dist
- [ ] 8. 修复构建/类型错误

## 验证
- `cd web && npm install && npm run build` 产出 `web/dist`（含 index.html + assets）。
- 后端 `make server` 后访问可加载 SPA（SPA fallback 由 Go 提供）。
