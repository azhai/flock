package main

import (
	"net/http"
	"testing"

	"github.com/azhai/flock/config"
	"github.com/azhai/flock/models"
)

// 关闭数据库后调用各 seed 函数，使每个内部 DB 操作都返回错误，
// 从而覆盖原本不可达的 db-error 分支（seedAdmin/seedCategories/seedDemo）。
// 结束时重新打开数据库，避免污染同一进程内的其它测试。
func TestSeedBranchesWithClosedDB(t *testing.T) {
	models.CloseDB()
	defer func() {
		if _, err := models.OpenDB(config.GetConfig().Database); err != nil {
			t.Fatalf("reopen db: %v", err)
		}
	}()

	// admin 配置非空（TestMain 已注入）→ 走向 SetPassword / Insert 失败分支
	seedAdmin()
	// 分类插入失败分支
	seedCategories()
	// 演示数据逐条失败分支
	seedDemo()
}

// TestClosedDBAPIHandlers 在关闭数据库后直接调用那些“查询失败即当作未命中、
// 继续执行写操作”的 handler，命中其 DB 写/读失败时的 500 分支。
// AuthMiddleware 仅校验 JWT 签名（不查库），故带合法 token 即可进入。
func TestClosedDBAPIHandlers(t *testing.T) {
	adminTok := authToken(t, "admin") // 需在关库前取 token
	models.CloseDB()
	defer func() {
		if _, err := models.OpenDB(config.GetConfig().Database); err != nil {
			t.Fatalf("reopen db: %v", err)
		}
	}()

	cases := []struct {
		name, method, path, token string
		body                      any
	}{
		{"list-categories", http.MethodGet, "/api/v1/categories", "", nil},
		{"create-category", http.MethodPost, "/api/v1/categories", adminTok, map[string]any{"name": "x"}},
		{"register-insert", http.MethodPost, "/api/v1/auth/register", "", map[string]any{"username": "zzz_new", "email": "zzz@flock.local", "password": "secret1"}},
		{"admin-list-users", http.MethodGet, "/api/v1/admin/users", adminTok, nil},
		{"admin-list-reports", http.MethodGet, "/api/v1/reports", adminTok, nil},
		{"list-tags", http.MethodGet, "/api/v1/tags", "", nil},
	}
	for _, c := range cases {
		code, b := do(t, c.method, c.path, c.token, c.body)
		if code != 500 {
			t.Errorf("%s: want 500, got %d (%s)", c.name, code, b)
		}
	}
}
