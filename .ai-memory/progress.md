# 进度日志

- 2026-09-15：完成全量源码阅读（18 个前端文件）；确认 construct-ui 为 Mithril 库；
  制定迁移计划 task_plan.md 与 findings.md。
- 已交付：脚手架（package.json/vite.config/index.html/styles.css/markdown.css）+ 核心 lib（api/store/notify/md/icon/ui/route）
  + 组件（Layout/Header/Sidebar/Pagination/MarkdownView/TagInput）+ 页面（Home/TopicDetail/NewTopic/Login/Register/
  UserProfile/Bookmarks/AdminPanel）+ main.ts（m.route）。删除全部旧 React 源码。
- `npm install` + `npm rebuild esbuild` + `npm run build` 通过，产出 web/dist（CSS 121KB 含 construct-ui，JS 1.1MB）。
- 已知取舍：TopicDetail 钓鱼线滚动条简化；表情选择器未移植；screenshot.png 仍为旧版截图。
