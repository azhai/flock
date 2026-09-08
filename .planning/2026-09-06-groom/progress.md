# Progress Log — flock

## 2026-09-06
- 调研：goent README(gitfolio models/handlers/middleware/helpers/config) → 已记录 findings.md
- 调研：echo/v5 版本确认 v5.3.1；环境 go1.27.1 + goproxy.cn
- 创建 `.planning/2026-09-06-flock/` 三文件 + `.ai-memory/adr/`
- **阶段 3**：go module init + echo/v5 API 签名确认（Context 结构体、GET/POST 签名、MiddlewareFunc、Group 中间件链）
- **阶段 4**：models 实体 User/Category/Topic/Post/PostLike + ForumSchema + AutoMigrate
  - 注意：goent `default:current_timestamp` 仅对 string 字段生效；CreatedAt/UpdatedAt 由应用层 time.Now() 填入
  - goent 插入时零值+default 字段会跳过让 DB 填值
- **阶段 5-6**：config + helpers(pagination/slug) + middleware(JWT auth/role)
- **阶段 7-8**：handlers(auth/category/topic/post/user/admin) + routes + main.go
  - echo v5 `GET(path, handler, middleware...)` 中间件必须在 handler 之后
  - `goent.Transaction` → 应使用 `model.Transaction`（goent 内部类型）
  - 修复：重复 package 行（环境钩子重复插入）、`err =` vs `err :=` 作用域、事务块重复
- **阶段 9**：go build/vet/gofmt 全绿
- **阶段 10-11**：前端 Vite+React+Chakra UI，npm run build 成功（499KB JS）
- **阶段 12**：`go run ./` 启动成功（8080），含/不含 .env 均可运行（NewEmptyEnviron fallback）

## 2026-09-07
- **设计优化 + 功能增强**（验收通过）：
  - 排版：`markdown.css` root 14→15px；修复 Header/Home 图标 `boxSize={14/12/11}`（≈49–56px）→ `boxSize={4}`（16px）；话题行/详情 py、gap、字号 xs→sm 加大留白
  - `\n` 换行：装 `remark-breaks`，`react-markdown` 加 `remarkGfm, remarkBreaks`
  - MD 编辑器工具栏：用已装 `@uiw/react-md-editor`（v4.1.2）替换 NewTopic 正文 Textarea 与 TopicDetail 回复/编辑框；样式仅 import `@uiw/react-md-editor/markdown-editor.css`（`esm/index.js` 自动引入 `./index.css`；`dist/mdeditor.css` 不在 exports 映射中会构建失败）；`data-color-mode` 按 prefers-color-scheme 适配
  - 全栈 tags：新表 `Tag`/`TopicTag`，Topic 挂载 `Tags []*Tag goe:"-"`；CreateTopic 支持 `tags[]`（upsertTag 去重复用）+ TopicTag 关联 + 计数；ListTopics `?tag=` 过滤 + 页内 loadTopicTags；GetTopic 带 tags；DeleteTopic 清理关联并回减计数；ListTags 仅返回 topics_count>0
  - 全栈 bookmark：新表 `Bookmark`(user_id+topic_id 联合唯一，镜像 PostLike，FK 不加 m2o 保唯一索引)；`POST /topics/:id/bookmark` 切换 + `GET /bookmarks`；TopicResponse 增 `bookmarked`
  - 帖子操作入口（后端原有，前端补齐）：TopicDetail 每楼 点赞/编辑(inline MDEditor)/删除；话题 置顶/锁定/删除(作者/staff)
  - 前端 api.ts 补齐 `listTags/bookmarkTopic/listBookmarks/updatePost/deletePost/toggleLike`，`listTopics` 加 `tag`、`createTopic` 加 `tags`
  - 新页 `Bookmarks.tsx`（我的收藏）+ 路由 `/bookmarks` + Header 下拉「我的收藏」
  - 验收：go build/vet/gofmt 全绿；`npm run build` 成功（1.59MB JS，含 MDEditor）；独立端口 8190 冒烟——tags 去重/过滤/收藏/收藏列表/删除级联清理/0计数过滤 全通过；浏览器可视化验证 无渲染错误/无 console 错误，编辑器工具栏可见，排版可读
- 遗留：8090 上运行的是旧二进制（旧 dist），需 `go build . ` 重建后重启才生效

## 2026-09-07 第二轮：通知/互动/审核 + 增强（后端完成）
- **模型**：新增 `Notification`/`TopicFollower`/`PostVote`/`Report` + `Topic.Followed`/`Post.VoteScore,MyVote` + `User.TopicsCount,PostsCount` carrier 挂载，均已进 ForumSchema（AutoMigrate 建表）
- **修复编译错误**：`User` 无 PostsCount → 加 carrier；`notificationSnapshot` 用 `.Filter().OrderBy()`（TableQuery 无 OrderBy）→ 改 `.Select()` 链；`Update().Exec()` 返回 1 值改 `err=`；SSE 用 `w.(http.Flusher)` 断言（echo v5 `c.Response()` 返回 `http.ResponseWriter`）；`@[\w\u4e00-\u9fa5]` 的 `\u` 在 Go regexp 非法 → 改 `[\p{L}\p{N}_-]+`
- **handlers**：`notification_handler.go`（notify/createReplyNotifications/createMentionNotifications/createLikeNotification/createReportNotifications + List/Unread/MarkRead/MarkAllRead/Stream SSE）、`vote_handler.go`（VotePost 顶/踩/改票/撤销 + removed 标志修正 myVote）、`report_handler.go`（CreateReport/ListReports/ResolveReport/DismissReport + deletePostInternal 级联）
- **删除级联**：`DeleteTopic` 增加 cleanup PostVote/Report/TopicFollower/Notification；`DeletePost` 复用 `deletePostInternal`
- **@提及首帖**：`CreateTopic` 补 `createMentionNotifications`（首帖含 @ 也会通知）
- **routes**：注册 /notifications(+/unread/read/read-all/stream)、/topics/:id/follow、/posts/:id/vote、/posts/:id/report、/reports(+resolve/dismiss)
- **验收**：go build/vet/gofmt 全绿。SSE 冒烟默认端口 8290（注意：本机 curl 被 http_proxy=127.0.0.1:10808 劫持 → 需 `export NO_PROXY=127.0.0.1`）：@提及→unread=1；关注+回复+@→unread=3；投票 顶→改-1→再点取消→cancel 语义全对（my_vote/score：1/1→-1/-1→0/0→0/0）；举报→resolve→帖子被级联删除（posts 由 2 归 1）
- 遗留：前端 f1-f4 未做（通知铃铛/评分举报 UI/emoji+热标签/AdminPanel 重做）

## 2026-09-07 第二轮收尾：前端 f1-f4 完成 + 端到端冒烟
- **前端 f1**：api.ts 增 votePost/reportPost/notifications/followTopic/reports/admin 接口；Header 铃铛下拉（未读红点 + `useNotifications` 钩子：fetch 流式 SSE 带 Bearer，断开回退 30s 轮询 + markRead/markAll）
- **前端 f2**：TopicDetail 再点取消的顶/踩评分、举报模态框、关注按钮、等级徽章、长帖右侧目录（`extractHeadings` 提取 MD 标题 `slugify` 锚点平滑滚动）
- **前端 f3**：NewTopic emoji 选择器（40 个常用表情 + Popover）、Home 左栏「热门标签」块（点击筛选）
- **前端 f4**：AdminPanel 重做 5 标签页（概览统计卡片/分类建删/举报流按状态筛选+处理删除违规帖+驳回/话题置顶锁定删除/用户搜索+角色+启停）
- **构建修复**：AdminPanel 第 216 行 `mb={3">` 引号闭合错误 → 修正；react-icons/fi 无 `FiPin` → 换 `FiStar`。`npm run build` 成功（1792 modules，JS 1.66MB）
- **冒烟修复**：ListReports `With("User","Post.User")` 嵌套语法不加载 Post → 改 `With("User","Post")`；`setReportStatus` 查举报也补 `With` 使返回带 user/post
- **验收**（独立 8190 + 干净 sqlite）：注册/登录 → 建话题带 tags+@bob（自动建 tag、等级 新手/初级）→ 回复@提及 → 顶/赞/关注/收藏 → 通知 reply+mention+like（unread 计数、单条/全部已读）→ 举报(带 post 内容) → resolve 级联删除违规回复帖(posts 2→1) → admin 置顶/锁定/用户列表/改角色/禁用（越权 alice 访问 admin/users=403、置顶=403 正确隔离）。go build/vet/gofmt 全绿。
- 遗留：生产 8090 跑旧二进制/旧 dist，需 `go build .` 重建并重启生效
