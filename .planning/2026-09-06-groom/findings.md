# Findings — flock 调研记录

## 2026-09-06 — goent ORM API
- **实体定义**：struct + `goe` tag：`pk`/`unique`/`index`/`default:...`/`column:...`/`m2o`
  外键、`o2m,fk:xxx` 反向、`goe:"-"` 挂载预加载字段。示例 `gitfolio/models/tables.go`。
- **表名**：默认 struct 名 snake_case（`User`→`user`，非复数）；可用 `TableName()` 覆盖。
- **Schema 层级**：`FooSchema struct{ User *goent.Table[User] }` + `goe:"public"` tag；
  `Database struct{ FooSchema; *goent.DB }`；`goent.Open[Database](drv)`。
- **迁移**：`goent.AutoMigrate(db)`；连接 `drivers.Connect(drivers.DatabaseConfig{Type,DSN,LogFile})`，
  `drivers.LoadConfig(env, defaultDSN)`。SQLite 驱动 `sqlite.OpenDSN`。
- **查询**：`db.T.Select().Where("id = ?", v).One()/.All()`；
  `.Filter(goent.Equals(db.T.Field("f"), v))`；`.Where("a = ? AND b = ?", ...)`；
  `.OrderBy("id DESC")`；分页 `.Skip(n).Take(n)` 或 `.Pagination(page,size)`；
  `.With("Field")` 预加载 m2o/o2m；聚合 `.Count("id")` 或 `db.T.Filter(...).Count("id")`。
- **写入**：`Insert().One(x)`、`Insert().All(autoincrement, slice)`；`Save().One(x)` 只更新非零值；
  `Update().Set(goent.Pair{Key,Value}).Filter(...).Exec()`；`Delete().Match(struct).Exec()`/`.Filter(...).Exec()`。
- **事务**：`db.BeginTransaction(func(tx goent.Transaction) error {...})`，查询/写 `.OnTransaction(tx)`。
- 参考：`/opt/repos/goent/README.md`、`/opt/repos/gitfolio/models/tables.go`、`handlers/*.go`。

## 2026-09-06 — gitfolio 后端模式（复用）
- **JWT**：`middleware/auth.go` Claims(UserID/Username/Email/Role) + `GenerateToken`；
  `c.Locals("user_id")` 存用户；`GetCurrentUserID/Role(ctx)` 读取；`AdminOnly/GuestReadOnly` 中间件。
- **DTO**：每实体 `XxxResponse` + `ToXxxResponse(model)` 手动转换（避免直接序列化 ORM 实体）。
- **helpers**：`GetPagination`、`Response{Data,Page,PerPage,Total}`、`NewPaginatedResponse`。
- **embed**：`//go:embed web/dist` + `fs.Sub` + SPA fallback 处理 `/*` 非 API 路由回退 index.html。
  （gitfolio 用 fiber；flock 改用 echo/v5 等价实现。）

## 2026-09-06 — 依赖版本与 replace
- echo/v5 可获取，最新 `v5.3.1`。go 1.27.1，GOPROXY=goproxy.cn。
- goent 本地 `/opt/repos/goent`(go 1.27, require gobus v0.2.1)；gobus 本地 `/opt/repos/gobus`。
- go.mod 需 `replace` 指向本地 goent/gobus，模式同 `gitfolio/go.mod`。

## Errors logged
- （暂无）
