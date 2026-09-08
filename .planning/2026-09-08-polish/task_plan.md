# 任务计划：SO 数据清洗 + 前端体验整备（2026-09-08）

## 目标
1. 清洗导入（HTML→Markdown）做成可复用 CMD 工具，并重导数据。含左下"作者+时间+评分"。
2. 高级分页组件，用于主题列表 + 管理后台用户列表（+管理话题）。
3. 顶/踩合并为一套（去掉冗余"赞"按钮），分数居中，精简帖子/主题工具栏。
4. 搜索：SQLite FTS5 全文检索（不引入 Redis / Meilisearch），接入右上角搜索框。

## 决策（已与用户确认）
- 评分模型：合并为一套顶/踩（PostVote 已是唯一来源，前端去掉"赞"按钮）。
- 缓存：不引入 Redis，数据库计数（PostVote 行即计数）+ 进程内足够。
- 搜索：SQLite FTS5（内建、零部署、无独立服务）。
- 正文：导入时用 html→markdown 转成 Markdown 后重导，与用户手写内容一致。

## 阶段
- [ ] A. cmd/import-so：正文 HTML→Markdown；重导数据（作者/时间/评分落位）。
- [ ] B. 高级分页：web/src/components/Pagination.jsx；供 Home 与 AdminPanel 使用。
- [ ] C. 顶/踩 UI 合并 + 工具栏精简（TopicDetail / PostCard / 主题头部）。
- [ ] E. FTS5 全文搜索：search.go 索引/重建 + ListTopics 使用 FTS；Header 搜索框接线。

## 验证
- 前端 `cd web && npm run build`。
- 后端 `go build ./... && go test ./...`。
- make run / dev 手动核对。