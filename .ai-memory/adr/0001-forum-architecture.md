# ADR 0001: flock 核心架构与数据模型

## Status
Accepted（前端技术栈部分已由 ADR 0004「Alpine.js 3」取代；其余架构决策仍有效）

## Context
用 Go 从零实现一个论坛（主参考 Discourse），需可单二进制部署（前端静态资源 embed）。
技术栈由需求固定：后端 echo/v5 + goent，前端 chakra-ui。首次交付聚焦「核心论坛闭环」。

## Decision
1. **数据库默认 SQLite**（goent 的 drivers 抽象可切 PostgreSQL）：零配置、符合 gitfolio
   单二进制部署惯例；`goent.AutoMigrate` 负责建表。
2. **认证用 JWT(HS256) + bcrypt**：复用 gitfolio 的 Claims/中间件模式；用户角色
   `admin` / `moderator` / `user`，未登录视作 `guest`（只读）。
3. **数据模型精简自 Discourse**：`User` / `Category`(parent_id 嵌套) / `Topic` /
   `Post`(post_number 楼号 + raw Markdown) / `Like`(post_id+user_id 唯一)。
   计数（views/posts/likes/reply）冗余存于 Topic 以支撑列表排序。
4. **前端 SPA**：Vite + React18 + Chakra UI，产物 `web/dist` 经 `//go:embed` 打包，
   echo 静态服务 + SPA fallback；Markdown 用 `@uiw/react-md-editor` 前端渲染，后端只存 raw。

## Consequences
- 优点：单二进制、起手零配置、复用 gitfolio 已验证的打包/鉴权范式。
- 天花板：SQLite 并发写入弱于 PG（可后续切）；Markdown 不服务端渲染（XSS 面由前端
  sanitize 控制，后续可加服务端 cooked + sanitize）；标签/通知/私信/全文搜索列为迭代项。
