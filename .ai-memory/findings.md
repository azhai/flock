# findings.md

## 前端产物 / .env（2026-09-07）
- `@uiw/react-md-editor` 默认版经 rehype-prism-plus + refractor 全量打包语言语法（~230K gzip）。
  改用其 `@uiw/react-md-editor/nohighlight` 子路径（`exports` 提供）可去掉 CodeMirror 与 prism，
  配合 `manualChunks` 拆包后所有 JS chunk gzip ≤100K（最大 react 85.7K）。
- environ: `NewEnv() == MustLoadEnvFile(".env")`（cwd 相对）；run() 已检测 `.env` 存在则用之，故后端在 make dev 下本就读 .env。
- vite.config.js 用 `loadEnv(mode, path.resolve(__dirname,'..'), '')` 从仓库根读 .env；代理目标取自 SERVER_PORT/BASE_URL（此前硬编码 8080 与 .env 的 8090 不符）。
- 详见 ADR 0002。

## 关键机制
- `environ.Lookup` → 失败时回退 `os.Getenv(prefix+key)`（见 environ.go:133）。因此测试用 `os.Setenv("DB_DSN", ...)` 即可被 `drivers.LoadConfig` 读中。
- `config.Load(env)` 全局缓存 `cfg`；同包 (package config) 测试可直接 `cfg=nil` 重置。
- `drivers.LoadConfig` 读 `DB_TYPE`(默认 sqlite) + `DB_DSN`(默认入参)。
- goent sqlite DSN 为文件路径；测试用 MkdirTemp 独立 DB。

## 覆盖率策略
- 主覆盖由 `main` 包集成测试（httptest + createApp）驱动，命中 handlers/middleware/routes/helpers/models/config/main。
- 纯函数分支在各自包内以单元测试补齐（trustLevel 的 >=50/>=10/…、To*Response(nil)、Slugify、GenerateToken/parseToken 等），避免用 API 造 50 个帖子。

## 难点
- StreamNotifications 无限循环 → 用可取消 context，取消后返回 nil，覆盖初始 send + Done 分支。
- 50 积分等级（资深）无法用 API 轻松达成 → 用 handlers/dto_test 直接测 trustLevel。