# Task Plan: groom 论坛软件（核心闭环 MVP）

## Goal
用 Go 实现一个可单二进制部署的论坛：用户注册/登录、嵌套分类、话题列表(排序/分页)、
话题帖子流、发帖/回复(Markdown)、点赞、用户主页、管理员(分类/用户/话题管理)，
前端 React+Chakra UI 静态文件 embed 进二进制，参考 ../gitfolio 的打包方式。

## 技术栈
- 后端：`github.com/labstack/echo/v5`(v5.3.1) + `github.com/azhai/goent`(本地 /opt/repos/goent)
- 数据库：SQLite（默认，可切 PostgreSQL）
- 认证：JWT(HS256) + bcrypt
- 前端：Vite + React18 + Chakra UI + react-router-dom，产物 `web/dist` 经 `//go:embed` 打包

## Phases
- [x] 1. 调研 goent CRUD/分页/事务/关联 API、gitfolio 鉴权+embed 模式 —— 已完成
- [x] 2. 计划三文件 + ADR —— 本阶段
- [ ] 3. 初始化 go module + 依赖(echo/v5、goent、gobus、jwt、bcrypt)，验证 echo/v5 API 签名
- [ ] 4. models 实体(User/Category/Topic/Post/Like) + Schema + AutoMigrate
- [ ] 5. config + helpers(分页、响应、slug 生成)
- [ ] 6. middleware(JWT auth + 角色权限)
- [ ] 7. handlers(auth/category/topic/post/user/admin)
- [ ] 8. routes + main.go + embed 静态文件
- [ ] 9. 后端闸门：gofmt + go build ./... + go vet ./...
- [ ] 10. 前端 web(Vite+React+Chakra UI)：页面 + API client
- [ ] 11. 前端 build，embed 打包，编译通过
- [ ] 12. 端到端：make run + 手动冒烟测试

## 数据模型（Discourse 精简）
- User: username/email 唯一、password(bcrypt)、role(admin/moderator/user)、active
- Category: name/slug 唯一、description、color、position、parent_id(嵌套, 可空)
- Topic: title、slug、category_id、user_id、pinned、closed、views/posts/likes 计数、last_posted_at
- Post: topic_id、user_id、post_number(楼号)、raw(Markdown)、reply_to_post_number
- Like: post_id+user_id 唯一（帖子点赞）

## 角色权限
- guest(未登录): 只读浏览
- user: 发帖/回复/点赞/编辑删除自己的帖子
- moderator: + 置顶/锁定话题、删除任意帖子
- admin: + 分类/用户/话题管理

## Open questions
- 数据库 SQLite vs PostgreSQL → 默认 SQLite（与 gitfolio 一致、零配置）
- 标签/收藏/通知/私信 → 列为后续迭代，不在 MVP

## Resume point
当前在阶段 3：初始化 go module。