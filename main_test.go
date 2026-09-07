package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/azhai/gobus/environ"
	"github.com/azhai/gouache/config"
	"github.com/azhai/gouache/middleware"
	"github.com/azhai/gouache/models"
	"github.com/labstack/echo/v5"
)

var testApp *echo.Echo

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gouache-test-*")
	if err != nil {
		fmt.Println("mkdirtemp:", err)
		os.Exit(1)
	}
	// environ.Lookup 回退 os.Getenv，直接用 Setenv 注入。
	os.Setenv("DB_DSN", filepath.Join(dir, "test.db"))
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("TOKEN_EXPIRY", "72")
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "adminpass")
	os.Setenv("ADMIN_EMAIL", "admin@gouache.local")

	cfg := config.Load(environ.NewEmptyEnviron())
	if _, err := models.OpenDB(cfg.Database); err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	seedAdmin()
	seedCategories()
	seedDemo()
	testApp = createApp()

	code := m.Run()

	models.CloseDB()
	os.RemoveAll(dir)
	os.Exit(code)
}

// authToken 为已有用户签发 token（覆盖 middleware.GenerateToken + config）。
func authToken(t *testing.T, username string) string {
	t.Helper()
	u, err := models.GetDB().User.Select().Where("username = ?", username).One()
	if err != nil || u == nil {
		t.Fatalf("user %q not found: %v", username, err)
	}
	tok, err := middleware.GenerateToken(u)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tok
}

func do(t *testing.T, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	return rawDo(t, method, path, token, "application/json", rd)
}

func rawDo(t *testing.T, method, path, token, ctype string, body io.Reader) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

func jmap(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json: %v (%s)", err, b)
	}
	return m
}

func jarr(t *testing.T, b []byte) []any {
	t.Helper()
	var a []any
	if err := json.Unmarshal(b, &a); err != nil && string(b) != "null" {
		t.Fatalf("json arr: %v (%s)", err, b)
	}
	return a
}

func num(m map[string]any, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func dataList(t *testing.T, b []byte) []any {
	t.Helper()
	m := jmap(t, b)
	if d, ok := m["data"].([]any); ok {
		return d
	}
	return nil
}

// TestFullIntegration 一条龙冒烟：覆盖绝大多数 handler 正常与异常分支。
func TestFullIntegration(t *testing.T) {
	adminTok := authToken(t, "admin")
	aliceTok := authToken(t, "alice")
	bobTok := authToken(t, "bob")

	// —— 注册 / 登录 ——
	var carolTok, malloryTok, mikeTok string
	t.Run("auth", func(t *testing.T) {
		code, b := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"username": "carol", "email": "carol@gouache.local", "password": "secret1"})
		if code != 201 {
			t.Fatalf("register carol: %d %s", code, b)
		}
		m := jmap(t, b)
		carolTok = m["token"].(string)

		// 重复用户名
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"username": "carol", "email": "carol2@gouache.local", "password": "secret1"}); code != 409 {
			t.Fatalf("dup username: %d", code)
		}
		// 邮箱已被注册（用 alice 邮箱）
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"username": "xx1", "email": "alice@gouache.local", "password": "secret1"}); code != 409 {
			t.Fatalf("dup email: %d", code)
		}
		// 用户名过短
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"username": "ab", "email": "ab@gouache.local", "password": "secret1"}); code != 400 {
			t.Fatalf("short username: %d", code)
		}
		// 邮箱非法
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"username": "abc", "email": "not-an-email", "password": "secret1"}); code != 400 {
			t.Fatalf("bad email: %d", code)
		}
		// 密码过短
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"username": "abc", "email": "abc@gouache.local", "password": "123"}); code != 400 {
			t.Fatalf("short pwd: %d", code)
		}
		// 非法 body
		if code, _ := rawDo(t, http.MethodPost, "/api/v1/auth/register", "", "application/json", strings.NewReader("{bad")); code != 400 {
			t.Fatalf("bad body: %d", code)
		}

		// 登录成功
		if code, b := do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"username": "carol", "password": "secret1"}); code != 200 {
			t.Fatalf("login: %d %s", code, b)
		}
		// 密码错误
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"username": "carol", "password": "wrong"}); code != 401 {
			t.Fatalf("wrong pwd: %d", code)
		}
		// 用户不存在
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"username": "ghost", "password": "x"}); code != 401 {
			t.Fatalf("ghost login: %d", code)
		}
		// 非法 body
		if code, _ := rawDo(t, http.MethodPost, "/api/v1/auth/login", "", "application/json", strings.NewReader("{bad")); code != 400 {
			t.Fatalf("login bad body: %d", code)
		}
	})

	// 注册第二个用户（作为“他人”验证权限）
	t.Run("register-more", func(t *testing.T) {
		for _, u := range []struct {
			user, email string
		}{{"mallory", "mallory@gouache.local"}, {"mike", "mike@gouache.local"}} {
			code, b := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
				"username": u.user, "email": u.email, "password": "secret1"})
			if code != 201 {
				t.Fatalf("register %s: %d", u.user, code)
			}
			_ = jmap(t, b)
		}
		malloryTok = authToken(t, "mallory")
		mikeTok = authToken(t, "mike")
	})

	// 把 mike 提升为版主（staff）
	t.Run("admin-promote", func(t *testing.T) {
		code, b := do(t, http.MethodPut, "/api/v1/admin/users/mike", adminTok, map[string]any{"role": "moderator"})
		if code != 200 {
			t.Fatalf("promote mike: %d %s", code, b)
		}
		// 角色已写入 JWT claims，需重新签发 token 让 StaffOnly 生效。
		mikeTok = authToken(t, "mike")
	})

	// —— 分类 ——
	var catID, catChildID int64
	t.Run("categories", func(t *testing.T) {
		// 列表（含 seed 的 general/tech）
		if code, b := do(t, http.MethodGet, "/api/v1/categories", "", nil); code != 200 {
			t.Fatalf("list cats: %d", code)
		} else if arr := jarr(t, b); len(arr) < 2 {
			t.Fatalf("cats too few: %d", len(arr))
		}

		// 创建父分类（admin）
		code, b := do(t, http.MethodPost, "/api/v1/categories", adminTok, map[string]any{
			"name": "测试分类", "slug": "tcat1", "description": "d", "color": "#f00", "position": 9})
		if code != 201 {
			t.Fatalf("create cat: %d %s", code, b)
		}
		catID = num(jmap(t, b), "id")

		// 创建子分类
		p := catID
		code, b = do(t, http.MethodPost, "/api/v1/categories", adminTok, map[string]any{
			"name": "子分类", "slug": "tcatc", "parent_id": p})
		if code != 201 {
			t.Fatalf("create child: %d", code)
		}
		catChildID = num(jmap(t, b), "id")

		// slug 已存在
		if code, _ := do(t, http.MethodPost, "/api/v1/categories", adminTok, map[string]any{
			"name": "测试分类2", "slug": "tcat1"}); code != 409 {
			t.Fatalf("dup slug: %d", code)
		}
		// 空名称
		if code, _ := do(t, http.MethodPost, "/api/v1/categories", adminTok, map[string]any{
			"name": "  ", "slug": "blank"}); code != 400 {
			t.Fatalf("blank name: %d", code)
		}
		// 非法 body
		if code, _ := rawDo(t, http.MethodPost, "/api/v1/categories", adminTok, "application/json", strings.NewReader("{bad")); code != 400 {
			t.Fatalf("bad body cat: %d", code)
		}
		// 非管理员创建
		if code, _ := do(t, http.MethodPost, "/api/v1/categories", bobTok, map[string]any{"name": "x"}); code == 200 {
			t.Fatalf("non-admin create should fail: %d", code)
		}

		// 更新子分类
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/categories/%d", catChildID), adminTok, map[string]any{
			"name": "子分类2", "description": "d2"}); code != 200 {
			t.Fatalf("update cat: %d", code)
		}
		// 自引用父分类
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/categories/%d", catChildID), adminTok, map[string]any{
			"parent_id": catChildID}); code != 400 {
			t.Fatalf("self parent: %d", code)
		}
		// 更新 slug 冲突
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/categories/%d", catChildID), adminTok, map[string]any{
			"slug": "tcat1"}); code != 409 {
			t.Fatalf("slug conflict update: %d", code)
		}
		// 更新不存在分类
		if code, _ := do(t, http.MethodPut, "/api/v1/categories/99999", adminTok, map[string]any{}); code != 404 {
			t.Fatalf("update missing cat: %d", code)
		}
		// 删除父分类（有子分类 → 400）
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%d", catID), adminTok, nil); code != 400 {
			t.Fatalf("delete parent with child: %d", code)
		}
	})

	// —— 标签 ——
	t.Run("tags", func(t *testing.T) {
		if code, b := do(t, http.MethodGet, "/api/v1/tags", "", nil); code != 200 {
			t.Fatalf("list tags: %d", code)
		} else if arr := jarr(t, b); len(arr) == 0 {
			t.Fatalf("no tags seeded: %d", len(arr))
		}
	})

	// —— 话题 创建 + 列表 ——
	var topicID int64
	t.Run("create-topic", func(t *testing.T) {
		code, b := do(t, http.MethodPost, "/api/v1/topics", carolTok, map[string]any{
			"title": "集成测试话题", "raw": "正文内容，@alice 你好。\n第二行。",
			"category_id": catID, "tags": []string{"集成", "go"}})
		if code != 201 {
			t.Fatalf("create topic: %d %s", code, b)
		}
		topicID = num(jmap(t, b), "id")
		if topicID <= 0 {
			t.Fatal("topic id invalid")
		}
		// 未登录
		if code, _ := do(t, http.MethodPost, "/api/v1/topics", "", map[string]any{
			"title": "x", "raw": "y"}); code != 401 {
			t.Fatalf("unauth create topic: %d", code)
		}
		// 空标题
		if code, _ := do(t, http.MethodPost, "/api/v1/topics", carolTok, map[string]any{
			"title": "  ", "raw": "y"}); code != 400 {
			t.Fatalf("empty title: %d", code)
		}
		// 空内容
		if code, _ := do(t, http.MethodPost, "/api/v1/topics", carolTok, map[string]any{
			"title": "x", "raw": "  "}); code != 400 {
			t.Fatalf("empty raw: %d", code)
		}
	})

	t.Run("list-topics", func(t *testing.T) {
		if code, b := do(t, http.MethodGet, "/api/v1/topics", "", nil); code != 200 {
			t.Fatalf("list topics: %d", code)
		} else if dl := dataList(t, b); len(dl) == 0 {
			t.Fatalf("no topics: %d", len(dl))
		}
		// 按分类过滤
		if code, _ := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics?category_id=%d", catID), "", nil); code != 200 {
			t.Fatalf("filter cat: %d", code)
		}
		// 关键词
		if code, b := do(t, http.MethodGet, "/api/v1/topics?q=集成", "", nil); code != 200 {
			t.Fatalf("filter q: %d", code)
		} else if dl := dataList(t, b); len(dl) != 1 {
			t.Fatalf("q should match 1: %d", len(dl))
		}
		// 标签过滤
		if code, b := do(t, http.MethodGet, "/api/v1/topics?tag=集成", "", nil); code != 200 {
			t.Fatalf("filter tag: %d", code)
		} else if dl := dataList(t, b); len(dl) != 1 {
			t.Fatalf("tag should match 1: %d", len(dl))
		}
		// 标签不存在 → 空
		if code, b := do(t, http.MethodGet, "/api/v1/topics?tag=nosuchtag", "", nil); code != 200 {
			t.Fatalf("no-such-tag: %d", code)
		} else if dl := dataList(t, b); len(dl) != 0 {
			t.Fatalf("no-such-tag should be empty: %d", len(dl))
		}
		// sort 变体
		for _, s := range []string{"latest", "top", "likes", "hot"} {
			if code, _ := do(t, http.MethodGet, "/api/v1/topics?sort="+s, "", nil); code != 200 {
				t.Fatalf("sort %s: %d", s, code)
			}
		}
		// 分页
		if code, b := do(t, http.MethodGet, "/api/v1/topics?page=1&per_page=2", "", nil); code != 200 {
			t.Fatalf("paginate: %d", code)
		} else if dl := dataList(t, b); len(dl) > 2 {
			t.Fatalf("per_page not respected: %d", len(dl))
		}
		// 登录用户访问（走 OptionalAuth 分支）
		if code, _ := do(t, http.MethodGet, "/api/v1/topics", carolTok, nil); code != 200 {
			t.Fatalf("list with auth: %d", code)
		}
	})

	var replyID int64
	t.Run("get-topic", func(t *testing.T) {
		if code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d", topicID), "", nil); code != 200 {
			t.Fatalf("get topic: %d", code)
		} else if num(jmap(t, b), "id") != topicID {
			t.Fatalf("get topic id mismatch")
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/topics/abc", "", nil); code != 400 {
			t.Fatalf("bad topic id: %d", code)
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/topics/999999", "", nil); code != 404 {
			t.Fatalf("missing topic: %d", code)
		}
	})

	t.Run("posts", func(t *testing.T) {
		// 列表
		if code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), "", nil); code != 200 {
			t.Fatalf("list posts: %d", code)
		} else if dl := dataList(t, b); len(dl) != 1 {
			t.Fatalf("posts should be 1: %d", len(dl))
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/topics/abc/posts", "", nil); code != 400 {
			t.Fatalf("bad posts topic id: %d", code)
		}

		// 回复
		code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), bobTok, map[string]any{
			"raw": "我来回复，@carol 看看。", "reply_to_post_number": 0})
		if code != 201 {
			t.Fatalf("create reply: %d %s", code, b)
		}
		replyID = num(jmap(t, b), "id")
		// 未登录
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), "", map[string]any{"raw": "x"}); code != 401 {
			t.Fatalf("unauth reply: %d", code)
		}
		// 空内容
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), bobTok, map[string]any{"raw": "  "}); code != 400 {
			t.Fatalf("empty reply: %d", code)
		}
		// 话题不存在
		if code, _ := do(t, http.MethodPost, "/api/v1/topics/999999/posts", bobTok, map[string]any{"raw": "x"}); code != 404 {
			t.Fatalf("reply missing topic: %d", code)
		}
		// 非法 body
		if code, _ := rawDo(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), bobTok, "application/json", strings.NewReader("{bad")); code != 400 {
			t.Fatalf("bad reply body: %d", code)
		}

		// 更新自己的回复
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", replyID), bobTok, map[string]any{"raw": "更新过的回复"}); code != 200 {
			t.Fatalf("update own post: %d", code)
		}
		// 他人更新 -> 403
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", replyID), carolTok, map[string]any{"raw": "x"}); code != 403 {
			t.Fatalf("update others post: %d", code)
		}
		// 更新不存在帖子
		if code, _ := do(t, http.MethodPut, "/api/v1/posts/999999", bobTok, map[string]any{"raw": "x"}); code != 404 {
			t.Fatalf("update missing post: %d", code)
		}
		// 空内容更新
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", replyID), bobTok, map[string]any{"raw": " "}); code != 400 {
			t.Fatalf("empty update: %d", code)
		}

		// 点赞后删除自己回复，再重建
		code, b = do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/like", replyID), carolTok, nil)
		if code != 200 {
			t.Fatalf("like: %d %s", code, b)
		} else if m := jmap(t, b); m["liked"] != true {
			t.Fatalf("like liked false: %v", m)
		}
		// 再点取消
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/like", replyID), carolTok, nil); code != 200 {
			t.Fatalf("unlike: %d", code)
		} else if m := jmap(t, b); m["liked"] != false {
			t.Fatalf("unlike liked false expected: %v", m)
		}
		// 点赞不存在帖子
		if code, _ := do(t, http.MethodPost, "/api/v1/posts/999999/like", carolTok, nil); code != 404 {
			t.Fatalf("like missing: %d", code)
		}
		// 未登录点赞
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/like", replyID), "", nil); code != 401 {
			t.Fatalf("unauth like: %d", code)
		}

		// 投票
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/vote", replyID), carolTok, map[string]any{"vote": 1}); code != 200 {
			t.Fatalf("vote +1: %d", code)
		} else if m := jmap(t, b); num(m, "score") != 1 || num(m, "my_vote") != 1 {
			t.Fatalf("vote +1 result: %v", m)
		}
		// 再投相同 → 取消
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/vote", replyID), carolTok, map[string]any{"vote": 1}); code != 200 {
			t.Fatalf("vote cancel: %d", code)
		} else if m := jmap(t, b); num(m, "score") != 0 {
			t.Fatalf("vote cancel result: %v", m)
		}
		// 非法投票值
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/vote", replyID), carolTok, map[string]any{"vote": 5}); code != 400 {
			t.Fatalf("illegal vote: %d", code)
		}
		// 投票不存在帖子
		if code, _ := do(t, http.MethodPost, "/api/v1/posts/999999/vote", carolTok, map[string]any{"vote": 1}); code != 404 {
			t.Fatalf("vote missing: %d", code)
		}

		// 删除首帖（post_number=1 → 400）
		first, _ := models.GetDB().Post.Select().
			Where("topic_id = ? AND post_number = ?", topicID, 1).One()
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", first.ID), carolTok, nil); code != 400 {
			t.Fatalf("delete first post: %d", code)
		}

		// 删除自己的回复
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", replyID), bobTok, nil); code != 200 {
			t.Fatalf("delete own post: %d", code)
		}
		// 重建回复供后续使用
		code, b = do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), carolTok, map[string]any{"raw": "新的回复"})
		if code != 201 {
			t.Fatalf("recreate reply: %d %s", code, b)
		}
		replyID = num(jmap(t, b), "id")
	})

	t.Run("posts-mine-votes", func(t *testing.T) {
		// 1) 已登录用户（无点赞/无投票）查看帖子流：命中 liked 批量查询与空循环
		if code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), carolTok, nil); code != 200 {
			t.Fatalf("authed list posts: %d", code)
		} else if len(dataList(t, b)) == 0 {
			t.Fatal("no posts in authed list")
		}
		// 2) carol 对回复投票后再次以 carol 身份查看 → 覆盖 loadPostVotes 的 mine 分支
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/vote", replyID), carolTok, map[string]any{"vote": 1}); code != 200 {
			t.Fatalf("vote before list: %d", code)
		}
		if code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), carolTok, nil); code != 200 {
			t.Fatalf("authed list after vote: %d", code)
		} else {
			posts := dataList(t, b)
			found := false
			for _, it := range posts {
				p := it.(map[string]any)
				if num(p, "id") == replyID && num(p, "my_vote") != 0 {
					found = true
				}
			}
			if !found {
				t.Fatalf("my_vote should be set for reply %d", replyID)
			}
		}
		// 3) 版主编辑他人回复（canEditPost staff 分支）→ 200
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", replyID), mikeTok, map[string]any{"raw": "版主代改"}); code != 200 {
			t.Fatalf("moderator edit others post: %d", code)
		}
		// 4) 帖子路由非法话题 id（CreatePost 的 ParseInt 分支）→ 400
		if code, _ := do(t, http.MethodPost, "/api/v1/topics/abc/posts", bobTok, map[string]any{"raw": "x"}); code != 400 {
			t.Fatalf("bad topic id in create post: %d", code)
		}
		// 5) 非法 id 的帖子删除（findPost ParseInt）→ 404
		if code, _ := do(t, http.MethodDelete, "/api/v1/posts/abc", bobTok, nil); code != 404 {
			t.Fatalf("bad id delete post: %d", code)
		}
		// 6) 非法 id 的帖子编辑 → 404
		if code, _ := do(t, http.MethodPut, "/api/v1/posts/abc", bobTok, map[string]any{"raw": "x"}); code != 404 {
			t.Fatalf("bad id update post: %d", code)
		}
	})

	t.Run("bookmark-follow", func(t *testing.T) {
		// 收藏开
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/bookmark", topicID), carolTok, nil); code != 200 {
			t.Fatalf("bookmark on: %d", code)
		} else if m := jmap(t, b); m["bookmarked"] != true {
			t.Fatalf("bookmark on: %v", m)
		}
		// 收藏列表（非空）→ 命中 loadBookmarks 主循环
		if code, b := do(t, http.MethodGet, "/api/v1/bookmarks", carolTok, nil); code != 200 {
			t.Fatalf("list bookmarks: %d", code)
		} else if a := jarr(t, b); len(a) != 1 {
			t.Fatalf("bookmarks should have 1: %d", len(a))
		}
		// 收藏关
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/bookmark", topicID), carolTok, nil); code != 200 {
			t.Fatalf("bookmark off: %d", code)
		} else if m := jmap(t, b); m["bookmarked"] != false {
			t.Fatalf("bookmark off: %v", m)
		}
		// 收藏列表（空）
		if code, b := do(t, http.MethodGet, "/api/v1/bookmarks", carolTok, nil); code != 200 {
			t.Fatalf("list bookmarks2: %d", code)
		} else if a := jarr(t, b); len(a) != 0 {
			t.Fatalf("bookmarks should be empty, got %d", len(a))
		}
		// 收藏不存在话题
		if code, _ := do(t, http.MethodPost, "/api/v1/topics/999999/bookmark", carolTok, nil); code != 404 {
			t.Fatalf("bookmark missing: %d", code)
		}
		// 记录一条收藏，供后续 /bookmarks 非空断言
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/bookmark", topicID), carolTok, nil); code != 200 {
			t.Fatalf("re-bookmark: %d", code)
		}
		// 关注开/关
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", topicID), carolTok, nil); code != 200 {
			t.Fatalf("follow on: %d", code)
		} else if m := jmap(t, b); m["followed"] != true {
			t.Fatalf("follow on: %v", m)
		}
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", topicID), carolTok, nil); code != 200 {
			t.Fatalf("follow off: %d", code)
		} else if m := jmap(t, b); m["followed"] != false {
			t.Fatalf("follow off: %v", m)
		}
	})

	t.Run("reports", func(t *testing.T) {
		// 举报回复
		code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", replyID), carolTok, map[string]any{"reason": "垃圾内容"})
		if code != 201 {
			t.Fatalf("report: %d %s", code, b)
		}
		rid := num(jmap(t, b), "id")
		// 重复举报（更新原因）
		if code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", replyID), carolTok, map[string]any{"reason": "再次举报"}); code != 200 {
			t.Fatalf("re-report: %d", code)
		} else if num(jmap(t, b), "id") != rid {
			t.Fatalf("re-report id mismatch")
		}
		// 空理由（handler 用 == "" 判断）
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", replyID), carolTok, map[string]any{"reason": ""}); code != 400 {
			t.Fatalf("empty reason: %d", code)
		}
		// 举报不存在帖子
		if code, _ := do(t, http.MethodPost, "/api/v1/posts/999999/report", carolTok, map[string]any{"reason": "x"}); code != 404 {
			t.Fatalf("report missing: %d", code)
		}
		// dismiss 新举报（在 resolve 删除回复之前做，避免目标帖子消失）
		code, b = do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", replyID), bobTok, map[string]any{"reason": "测试"})
		if code != 201 {
			t.Fatalf("report2: %d", code)
		}
		rid2 := num(jmap(t, b), "id")
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/dismiss", rid2), adminTok, nil); code != 200 {
			t.Fatalf("dismiss: %d", code)
		}
		// 受理一个已有相同 status 的举报（走 status==status 分支）
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/dismiss", rid2), adminTok, nil); code != 200 {
			t.Fatalf("dismiss repeat: %d", code)
		}
		// 非法 id
		if code, _ := do(t, http.MethodPost, "/api/v1/reports/abc/dismiss", adminTok, nil); code != 400 {
			t.Fatalf("bad report id: %d", code)
		}
		// 管理员查看举报流
		if code, b := do(t, http.MethodGet, "/api/v1/reports?status=pending", adminTok, nil); code != 200 {
			t.Fatalf("list reports: %d", code)
		} else if len(dataList(t, b)) == 0 {
			t.Fatalf("no reports")
		}
		// 处理 → 删除回复
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/resolve", rid), adminTok, nil); code != 200 {
			t.Fatalf("resolve: %d", code)
		}
		// 处理不存在举报
		if code, _ := do(t, http.MethodPost, "/api/v1/reports/999999/resolve", adminTok, nil); code != 404 {
			t.Fatalf("resolve missing: %d", code)
		}
	})

	t.Run("notifications", func(t *testing.T) {
		// alice 是众多 seed 话题作者 + 被 @提及 → 有通知
		if code, b := do(t, http.MethodGet, "/api/v1/notifications", aliceTok, nil); code != 200 {
			t.Fatalf("list notif: %d", code)
		} else if len(jarr(t, b)) == 0 {
			t.Fatalf("alice should have notifications")
		}
		// 未登录
		if code, _ := do(t, http.MethodGet, "/api/v1/notifications", "", nil); code != 401 {
			t.Fatalf("unauth notif: %d", code)
		}
		// 未读数
		code, b := do(t, http.MethodGet, "/api/v1/notifications/unread", aliceTok, nil)
		if code != 200 {
			t.Fatalf("unread count: %d", code)
		}
		unread := num(jmap(t, b), "unread")
		if unread == 0 {
			t.Fatalf("unread should be >0")
		}
		// 标记单条已读（取第一条 id）
		list := jarr(t, fetch(t, aliceTok, "/api/v1/notifications"))
		firstID := num(list[0].(map[string]any), "id")
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/notifications/%d/read", firstID), aliceTok, nil); code != 200 {
			t.Fatalf("mark read: %d", code)
		}
		// 标记全部已读
		if code, _ := do(t, http.MethodPost, "/api/v1/notifications/read-all", aliceTok, nil); code != 200 {
			t.Fatalf("read-all: %d", code)
		}
		if code, b := do(t, http.MethodGet, "/api/v1/notifications/unread", aliceTok, nil); code != 200 {
			t.Fatalf("unread after: %d", code)
		} else if num(jmap(t, b), "unread") != 0 {
			t.Fatalf("unread should be 0")
		}
		// mark read 非法 id
		if code, _ := do(t, http.MethodPost, "/api/v1/notifications/abc/read", aliceTok, nil); code != 400 {
			t.Fatalf("mark read bad id: %d", code)
		}
	})

	t.Run("notifications-stream", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/stream", nil).WithContext(ctx)
		req.Header.Set("Authorization", "Bearer "+aliceTok)
		w := httptest.NewRecorder()
		done := make(chan bool)
		go func() { testApp.ServeHTTP(w, req); done <- true }()
		time.Sleep(120 * time.Millisecond)
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("stream did not stop on cancel")
		}
		if !strings.Contains(w.Body.String(), "event: notify") {
			t.Fatalf("stream missing event: %q", w.Body.String())
		}
	})

	t.Run("users", func(t *testing.T) {
		if code, b := do(t, http.MethodGet, "/api/v1/users/alice", "", nil); code != 200 {
			t.Fatalf("get user: %d", code)
		} else if m := jmap(t, b); m["username"] != "alice" {
			t.Fatalf("get user name: %v", m)
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/users/nobody", "", nil); code != 404 {
			t.Fatalf("get missing user: %d", code)
		}
		// 用户话题 / 帖子
		if code, _ := do(t, http.MethodGet, "/api/v1/users/alice/topics", "", nil); code != 200 {
			t.Fatalf("user topics: %d", code)
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/users/alice/posts", "", nil); code != 200 {
			t.Fatalf("user posts: %d", code)
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/users/nobody/topics", "", nil); code != 404 {
			t.Fatalf("missing user topics: %d", code)
		}
		if code, _ := do(t, http.MethodGet, "/api/v1/users/nobody/posts", "", nil); code != 404 {
			t.Fatalf("missing user posts: %d", code)
		}
		// me
		if code, b := do(t, http.MethodGet, "/api/v1/auth/me", carolTok, nil); code != 200 {
			t.Fatalf("me: %d", code)
		} else if m := jmap(t, b); m["username"] != "carol" {
			t.Fatalf("me username: %v", m)
		}
		// 未登录 me
		if code, _ := do(t, http.MethodGet, "/api/v1/auth/me", "", nil); code != 401 {
			t.Fatalf("unauth me: %d", code)
		}
		// 更新资料
		if code, _ := do(t, http.MethodPut, "/api/v1/user/me", carolTok, map[string]any{
			"full_name": "Carol", "bio": "hi", "website": "https://x.io", "location": "sh"}); code != 200 {
			t.Fatalf("update profile: %d", code)
		}
		if code, _ := do(t, http.MethodPut, "/api/v1/user/me", "", map[string]any{}); code != 401 {
			t.Fatalf("unauth update profile: %d", code)
		}
	})

	t.Run("admin-users", func(t *testing.T) {
		if code, b := do(t, http.MethodGet, "/api/v1/admin/users?q=car", adminTok, nil); code != 200 {
			t.Fatalf("list users: %d", code)
		} else if len(dataList(t, b)) == 0 {
			t.Fatalf("no users in list")
		}
		// 更新普通用户
		if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/carol", adminTok, map[string]any{
			"full_name": "CC", "role": "user"}); code != 200 {
			t.Fatalf("admin update user: %d", code)
		}
		// 非法角色
		if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/carol", adminTok, map[string]any{"role": "superuser"}); code != 400 {
			t.Fatalf("bad role: %d", code)
		}
		// 密码过短
		if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/carol", adminTok, map[string]any{"password": "123"}); code != 400 {
			t.Fatalf("short pwd admin: %d", code)
		}
		// 更新 is_active / bio / full_name（覆盖 AdminUpdateUser 更多分支）
		active := false
		if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/carol", adminTok, map[string]any{
			"full_name": "CC2", "bio": "测试简介", "is_active": active}); code != 200 {
			t.Fatalf("admin update is_active: %d", code)
		}
		// 禁用账号登录 → 403（Login 的 IsActive 分支）
		if code, _ := do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"username": "carol", "password": "secret1"}); code != 403 {
			t.Fatalf("inactive login want 403: %d", code)
		}
		// 恢复启用，便于后续用例继续使用 carol 身份
		reactivate := true
		if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/carol", adminTok, map[string]any{"is_active": reactivate}); code != 200 {
			t.Fatalf("reactivate carol: %d", code)
		}
		// 用户不存在
		if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/nobody", adminTok, map[string]any{}); code != 404 {
			t.Fatalf("admin update missing: %d", code)
		}
		// 非管理员访问
		if code, _ := do(t, http.MethodGet, "/api/v1/admin/users", bobTok, nil); code == 200 {
			t.Fatalf("non-admin list users should fail: %d", code)
		}
	})

	t.Run("moderation", func(t *testing.T) {
		// 非 staff（bob）置顶 → 403
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/pin", topicID), bobTok, nil); code != 403 {
			t.Fatalf("non-staff pin: %d", code)
		}
		// 版主置顶/取消
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/pin", topicID), mikeTok, nil); code != 200 {
			t.Fatalf("pin: %d", code)
		}
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/unpin", topicID), mikeTok, nil); code != 200 {
			t.Fatalf("unpin: %d", code)
		}
		// 关闭 → 回复 403 → 打开
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/close", topicID), mikeTok, nil); code != 200 {
			t.Fatalf("close: %d", code)
		}
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), bobTok, map[string]any{"raw": "被锁"}); code != 403 {
			t.Fatalf("reply to closed: %d", code)
		}
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/open", topicID), mikeTok, nil); code != 200 {
			t.Fatalf("open: %d", code)
		}
		// 操作不存在话题
		if code, _ := do(t, http.MethodPost, "/api/v1/topics/999999/pin", mikeTok, nil); code != 404 {
			t.Fatalf("pin missing: %d", code)
		}
	})

	t.Run("topic-edit-delete", func(t *testing.T) {
		// 作者更新话题
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/topics/%d", topicID), carolTok, map[string]any{
			"title": "集成测试话题v2"}); code != 200 {
			t.Fatalf("update own topic: %d", code)
		}
		// 版主代改他人话题（canEditTopic staff 分支）→ 200
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/topics/%d", topicID), mikeTok, map[string]any{
			"title": "版主代改话题"}); code != 200 {
			t.Fatalf("moderator update others topic: %d", code)
		}
		// 他人更新 → 403
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/topics/%d", topicID), malloryTok, map[string]any{
			"title": "x"}); code != 403 {
			t.Fatalf("update others topic: %d", code)
		}
		// 更新不存在话题
		if code, _ := do(t, http.MethodPut, "/api/v1/topics/999999", carolTok, map[string]any{}); code != 404 {
			t.Fatalf("update missing topic: %d", code)
		}
		// 空标题保留原值（200）
		if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/topics/%d", topicID), carolTok, map[string]any{
			"title": "  "}); code != 200 {
			t.Fatalf("blank title keep: %d", code)
		}
		// 他人删除 → 403
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/topics/%d", topicID), malloryTok, nil); code != 403 {
			t.Fatalf("delete others topic: %d", code)
		}
		// 删除不存在
		if code, _ := do(t, http.MethodDelete, "/api/v1/topics/999999", carolTok, nil); code != 404 {
			t.Fatalf("delete missing topic: %d", code)
		}
		// 作者删除（级联清理）→ 200
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/topics/%d", topicID), carolTok, nil); code != 200 {
			t.Fatalf("delete own topic: %d", code)
		}
		// 删除后确认消失
		if code, _ := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d", topicID), "", nil); code != 404 {
			t.Fatalf("topic not gone: %d", code)
		}
	})

	t.Run("admin-delete-cat", func(t *testing.T) {
		// 有话题的分类删除 → 400（tech 有 seed 话题）
		tech, _ := models.GetDB().Category.Select().Where("slug = ?", "tech").One()
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%d", tech.ID), adminTok, nil); code != 400 {
			t.Fatalf("delete non-empty cat: %d", code)
		}
		// 删除子分类（本次未挂话题且无子）→ 200
		if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%d", catChildID), adminTok, nil); code != 200 {
			t.Fatalf("delete cat child: %d", code)
		}
		// 删除不存在分类
		if code, _ := do(t, http.MethodDelete, "/api/v1/categories/999999", adminTok, nil); code != 404 {
			t.Fatalf("delete missing cat: %d", code)
		}
	})

	t.Run("spa", func(t *testing.T) {
		if code, b := rawDo(t, http.MethodGet, "/", "", "", nil); code != 200 {
			t.Fatalf("SPA index: %d", code)
		} else if !strings.Contains(string(b), "gouache") {
			t.Fatalf("index.html missing brand: %s", b)
		}
		if code, _ := rawDo(t, http.MethodGet, "/api/nope", "", "", nil); code != 404 {
			t.Fatalf("unknown api: %d", code)
		}
	})

	// 确保使用了相关 token，避免 unused
	_, _, _ = carolTok, malloryTok, mikeTok
	_, _ = aliceTok, bobTok
}

// TestMeGhostAndFollow 覆盖 Me 的“用户不存在”分支，以及 loadFollowed 的已关注非空分支。
func TestMeGhostAndFollow(t *testing.T) {
	// 幽灵用户：claims 存在但数据库无此用户 → GetCurrentUser 返回 nil → /auth/me 401
	ghost, err := middleware.GenerateToken(&models.User{
		ID: 999999, Username: "ghost", Email: "ghost@gouache.local", Role: models.RoleUser})
	if err != nil {
		t.Fatalf("ghost token: %v", err)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/auth/me", ghost, nil); code != 401 {
		t.Fatalf("ghost me want 401: %d", code)
	}

	// loadFollowed：让 admin 关注一个话题后，以 admin 登录读取该话题 → 触发关注列表非空。
	adminTokN := authToken(t, "admin")
	topic, _ := models.GetDB().Topic.Select().One()
	if topic == nil {
		t.Fatal("no topic in db")
	}
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", topic.ID), adminTokN, nil); code != 200 {
		t.Fatalf("follow before get: %d", code)
	}
	if code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d", topic.ID), adminTokN, nil); code != 200 {
		t.Fatalf("get followed topic: %d", code)
	} else if m := jmap(t, b); m["followed"] != true {
		t.Fatalf("followed should be true: %s", b)
	}
}

func fetch(t *testing.T, token, path string) []byte {
	t.Helper()
	code, b := do(t, http.MethodGet, path, token, nil)
	if code != 200 {
		t.Fatalf("fetch %s: %d", path, code)
	}
	return b
}
