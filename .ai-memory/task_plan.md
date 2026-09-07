# task_plan.md — 95%+ 覆盖率 + Makefile 热重载测试

## 目标
1. 整个项目 Go 测试覆盖率 ≥ 95%（`go test -coverpkg=./... ./...` 聚合）。
2. 前端 vitest 测试搭建（合理覆盖）。
3. Makefile 增加：编译 + 测试热重载目标。

## 阶段
- [x] S1 勘察：已读完 handlers/middleware/models/config/helpers/routes/main。
- [ ] S2 测试基建：TestMain(env+sqlite+seed) + httptest helper。
- [ ] S3 后端测试：
  - [ ] main_test.go（集成：全 API 正/负路径）
  - [ ] handlers/dto_test.go（纯函数分支）
  - [ ] helpers/helpers_test.go
  - [ ] config/config_test.go
  - [ ] middleware/auth_test.go
  - [ ] models/models_test.go
- [ ] S4 覆盖率迭代到 ≥95%。
- [ ] S5 前端 vitest 搭建 + 抽样测试。
- [ ] S6 Makefile：test / test-cover / test-watch。

## 验证
- `go test -coverpkg=./... ./... -coverprofile=/tmp/c.out && go tool cover -func=/tmp/c.out | tail -1`
- `cd web && npx vitest run`

## 关键发现（详见 findings.md）
- environ.Lookup 回退到 os.Getenv（含 KeyPrefix），故 TestMain 用 os.Setenv 注入 DB_DSN 等即可。
- config.Load 有全局 cfg 缓存 → 同包 test 内可直接重置。
- goent 使用 sqlite，走 DB_DSN。