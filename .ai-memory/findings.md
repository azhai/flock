# 调研发现

## Construct UI 是 Mithril 库（非 React）
- 组件：Button, Input, Card, Dialog, Drawer, Tabs/TabItem, Toast, Form, Select, Tag Input,
  Tag, Tree, Breadcrumbs, Callout, EmptyState, Tooltip, Table, Menu, Popover, Collapse, Icon, Spinner。
- Utils: Focus Manager, Responsive Manager, Transition Manager。
- 引入 CSS：`construct-ui/lib/index.css`。
- 不支持内建主题/暗黑模式（仅固定 Colors 文档），配色需自行用 CSS 控制。
- 图标：依赖独立图标包（如 `construct-ui` 自带的 Icons 或需自行提供）。
  → 项目改用 `feather-icons` 风格的 SVG 或 `react-icons` 不行（React）。改用纯 SVG / unicode / 自建小图标组件。
  决定：用一个轻量 `Icon` 组件 + 内联 SVG（feather 路径）覆盖所需图标。

## API 形状（来自源码阅读）
- api.ts 完全可复用（fetch 封装，Bearer 从 localStorage）。
- 后端返回结构：listTopics 返回 `{ data: [...], total }`；getTopic 返回 topic 对象（含 tags/category/user）；
  listPosts 返回 `{ data, total }`；login 返回 `{ user, token }`；notifications 为数组；
  listReports/listUsers/listTopics(admin) 返回 `{ data, total }`。
- 帖子分数：post 带 `like_count`、`my_vote`、`score`(投票分)；topic 带 `like_count`、`views_count`、`posts_count`、`bookmarked`、`followed`、`pinned`、`closed`、`hidden`。

## 简化项（迁移取舍）
- TopicDetail 钓鱼线滚动条+吸附：改为普通 `overflow:auto` 容器 + 右侧帖子导航点（点击跳转），保留形态。
- 表情选择器（Popover）：用简单下拉或省略，优先功能。
- 头像：用文字首字母圆形（原 Chakra Avatar name=username），无需图片。
