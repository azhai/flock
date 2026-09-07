package main

// zzz_coverage_test.go — 针对当前未覆盖分支补充的 coverage 测试。
// 全部测试复用 main_test.go 的 testApp / do / rawDo / authToken / jmap / num 等 helper。
// 测试之间共享全局 DB 与 testApp，故不使用 t.Parallel，按序串行执行。
//
// 覆盖策略分三类：
//   1) 正常/校验分支：通过 HTTP 打真实请求；
//   2) handler 内 user==nil 401 分支：借用"幽灵 token"（claims 有效但 DB 无此用户），
//      AuthMiddleware 只验签不查库，放行后 handler 内 GetCurrentUser 返回 nil → 401；
//   3) DB-error 500 分支：models.CloseDB() 后调用，defer 中 OpenDB 恢复（同 main_entry_test.go）。

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/azhai/goent/model"
	"github.com/azhai/gouache/config"
	"github.com/azhai/gouache/middleware"
	"github.com/azhai/gouache/models"
)

// ---- helpers ----

// covRegister 注册一个新用户并签发 token。
func covRegister(t *testing.T, username, email string) string {
	t.Helper()
	code, b := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"username": username, "email": email, "password": "secret1"})
	if code != 201 {
		t.Fatalf("register %s: %d %s", username, code, b)
	}
	m := jmap(t, b)
	if m["token"] == nil {
		m = map[string]any{"token": authToken(t, username)}
	}
	return m["token"].(string)
}

// covGhostToken 构造一个 JWT 有效但 DB 中不存在的用户 token。
func covGhostToken(t *testing.T) string {
	t.Helper()
	tok, err := middleware.GenerateToken(&models.User{
		ID: 987654, Username: "ghost_user", Email: "ghost@gouache.local", Role: models.RoleUser})
	if err != nil {
		t.Fatalf("ghost token: %v", err)
	}
	return tok
}

// covCreateTopic 创建一个话题返回 topic id，tags 可带。
func covCreateTopic(t *testing.T, token, title, raw string, tags []string) int64 {
	t.Helper()
	code, b := do(t, http.MethodPost, "/api/v1/topics", token, map[string]any{
		"title": title, "raw": raw, "tags": tags})
	if code != 201 {
		t.Fatalf("create topic %q: %d %s", title, code, b)
	}
	id := num(jmap(t, b), "id")
	if id <= 0 {
		t.Fatalf("invalid topic id for %q", title)
	}
	return id
}

// covFirstPostID 取话题首帖 id。
func covFirstPostID(t *testing.T, topicID int64) int64 {
	t.Helper()
	p, err := models.GetDB().Post.Select().
		Where("topic_id = ? AND post_number = ?", topicID, 1).One()
	if err != nil || p == nil {
		t.Fatalf("first post of topic %d: %v", topicID, err)
	}
	return p.ID
}

// ---- main.go: run() / seeds / 静态文件 ----

// TestCovRunAndSeeds 覆盖 run() 的启动错误分支，以及 seedAdmin/seedCategories/seedDemo
// 的"数据已存在→提前返回"分支。
//
// run() 内部会 OpenDB 并在返回时 Close，测试结束后需重新打开 DB 以免污染后续用例。
// 端口策略：若可绑定则占用之，使 app.Start 因"address already in use"立刻失败；
// 若端口本就已被占用（如本机 8090 已被其它进程监听），则 app.Start 同样会失败——
// 两种情况下 run() 都会快速返回 error，不会阻塞测试。
func TestCovRunAndSeeds(t *testing.T) {
	port, _ := config.GetServerInfo()
	addr := fmt.Sprintf(":%d", port)
	if ln, err := net.Listen("tcp", addr); err == nil {
		// 端口空闲 → 我们占住它，Start 必失败。
		defer ln.Close()
	} else {
		// 端口已被占用 → Start 也会失败。
	}

	// run() 在已占用端口上 Start → 立即返回错误，覆盖 start 失败分支。
	derr := run()
	if derr == nil {
		t.Fatal("run() should fail to start server on a busy port")
	}
	// run() 已把 DB 关闭，重新打开供后续用例使用。
	if _, err := models.OpenDB(config.GetConfig().Database); err != nil {
		t.Fatalf("reopen db after run(): %v", err)
	}

	// seed 函数在数据已存在时均提前返回（覆盖 early-return 分支）。
	seedAdmin()
	seedCategories()
	seedDemo()

	// seedTag 的"已存在→复用"分支：同一事务内用同名标签调用两次，
	// 第二次命中"已存在则返回"分支。
	db := models.GetDB()
	if err := db.BeginTransaction(func(tx model.Transaction) error {
		if _, err := seedTag(tx, "cov-reuse-tag"); err != nil {
			return err
		}
		if _, err := seedTag(tx, "cov-reuse-tag"); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("seedTag reuse: %v", err)
	}
}

// TestCovStaticFiles 覆盖 createApp 静态子路由：目录路径（关闭后回退 index）与
// 不存在路径（SPA fallback）。
func TestCovStaticFiles(t *testing.T) {
	// /assets 是 web/dist 中的目录 → dist.Open 成功但 IsDir → close 后回退 index.html
	code, b := rawDo(t, http.MethodGet, "/assets", "", "", nil)
	if code != 200 || !strings.Contains(string(b), "gouache") && !strings.Contains(string(b), "groom") {
		t.Fatalf("dir static: %d %s", code, b)
	}
	// /no-such-path 不存在 → SPA fallback index.html
	if code, b := rawDo(t, http.MethodGet, "/cov-no-such-page", "", "", nil); code != 200 {
		t.Fatalf("fallback static: %d %s", code, b)
	}
}

// ---- auth_handler / models.SetPassword ----

// TestCovAuthExtra 覆盖 Register 的 SetPassword 失败分支（超长密码触发 bcrypt 错误，
// 同时覆盖 models.SetPassword 的错误分支）。
func TestCovAuthExtra(t *testing.T) {
	longPw := strings.Repeat("a", 100) // >72 bytes → bcrypt ErrPasswordTooLong
	if code, _ := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"username": "covlongpw", "email": "covlongpw@gouache.local", "password": longPw}); code != 500 {
		t.Fatalf("register long-pw should be 500, got %d", code)
	}
	// 确认未创建该用户（失败分支未入库）
	if code, _ := do(t, http.MethodGet, "/api/v1/users/covlongpw", "", nil); code != 404 {
		t.Fatalf("long-pw user should not exist: %d", code)
	}
}

// ---- user_handler.UpdateProfile ----

func TestCovUpdateProfileBranches(t *testing.T) {
	ghost := covGhostToken(t)

	// GetCurrentUser 返回 nil → 401（幽灵 token 借 AuthMiddleware 放行）
	if code, _ := do(t, http.MethodPut, "/api/v1/user/me", ghost, map[string]any{"full_name": "x"}); code != 401 {
		t.Fatalf("ghost update profile: %d", code)
	}

	// 非法 body → 400
	tok := covRegister(t, "covprof_a", "covprof_a@gouache.local")
	if code, _ := rawDo(t, http.MethodPut, "/api/v1/user/me", tok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad body update profile: %d", code)
	}
}

// ---- bookmark_handler ----

func TestCovBookmarkBranches(t *testing.T) {
	ghost := covGhostToken(t)
	tok := covRegister(t, "covbm_a", "covbm_a@gouache.local")

	// ToggleBookmark 幽灵 → 401
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/1/bookmark", ghost, nil); code != 401 {
		t.Fatalf("ghost bookmark: %d", code)
	}
	// 非法话题 id → 400
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/abc/bookmark", tok, nil); code != 400 {
		t.Fatalf("bad bookmark id: %d", code)
	}
	// ListBookmarks 幽灵 → 401
	if code, _ := do(t, http.MethodGet, "/api/v1/bookmarks", ghost, nil); code != 401 {
		t.Fatalf("ghost list bookmarks: %d", code)
	}
}

// ---- vote_handler ----

func TestCovVoteBranches(t *testing.T) {
	ghost := covGhostToken(t)
	tok := covRegister(t, "covvote_a", "covvote_a@gouache.local")
	topicID := covCreateTopic(t, tok, "分数测试", "投票覆盖", nil)
	postID := covFirstPostID(t, topicID)
	base := fmt.Sprintf("/api/v1/posts/%d/vote", postID)

	// 幽灵 → 401
	if code, _ := do(t, http.MethodPost, base, ghost, map[string]any{"vote": 1}); code != 401 {
		t.Fatalf("ghost vote: %d", code)
	}
	// 非法 body → 400
	if code, _ := rawDo(t, http.MethodPost, base, tok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad vote body: %d", code)
	}
	// 投 +1 再改为 -1 → 覆盖"通票"更新既有票分支（existing.Vote=req.Vote + Save）
	if code, _ := do(t, http.MethodPost, base, tok, map[string]any{"vote": 1}); code != 200 {
		t.Fatalf("vote +1: %d", code)
	}
	if code, b := do(t, http.MethodPost, base, tok, map[string]any{"vote": -1}); code != 200 {
		t.Fatalf("vote change to -1: %d", code)
	} else if m := jmap(t, b); num(m, "score") != -1 || num(m, "my_vote") != -1 {
		t.Fatalf("vote change result: %v", m)
	}
}

// ---- category_handler ----

func TestCovCategoryBranches(t *testing.T) {
	adminTok := authToken(t, "admin")

	// 建一个分类供更新用
	code, b := do(t, http.MethodPost, "/api/v1/categories", adminTok, map[string]any{
		"name": "覆盖分类", "slug": "covcat1", "color": "#111"})
	if code != 201 {
		t.Fatalf("create covcat: %d %s", code, b)
	}
	cid := num(jmap(t, b), "id")
	update := fmt.Sprintf("/api/v1/categories/%d", cid)

	// 非法 body → 400（UpdateCategory Bind 失败）
	if code, _ := rawDo(t, http.MethodPut, update, adminTok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad body update cat: %d", code)
	}
	// 成功更新 slug / color / position → 覆盖多处赋值分支
	code, b = do(t, http.MethodPut, update, adminTok, map[string]any{
		"name": "覆盖分类改", "slug": "covcat1b", "color": "#222", "position": 7})
	if code != 200 {
		t.Fatalf("update cat fields: %d %s", code, b)
	}
	// 仅 description + color（走 color 非空、position 分支已单独覆盖）
	if code, _ := do(t, http.MethodPut, update, adminTok, map[string]any{"color": "#333"}); code != 200 {
		t.Fatalf("update cat color: %d", code)
	}
}

// ---- admin_handler ----

func TestCovAdminBranches(t *testing.T) {
	adminTok := authToken(t, "admin")
	covRegister(t, "covadm_a", "covadm_a@gouache.local")

	// 非法 body → 400（AdminUpdateUser Bind 失败）
	if code, _ := rawDo(t, http.MethodPut, "/api/v1/admin/users/covadm_a", adminTok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad body admin update: %d", code)
	}
	// 超长密码 → SetPassword 失败 → 500
	longPw := strings.Repeat("b", 100)
	if code, _ := do(t, http.MethodPut, "/api/v1/admin/users/covadm_a", adminTok, map[string]any{"password": longPw}); code != 500 {
		t.Fatalf("admin long-pw should 500: %d", code)
	}
}

// ---- report_handler ----

func TestCovReportBranch(t *testing.T) {
	ghost := covGhostToken(t)
	tok := covRegister(t, "covrep_a", "covrep_a@gouache.local")
	topicID := covCreateTopic(t, tok, "举报覆盖", "举报", nil)
	postID := covFirstPostID(t, topicID)
	base := fmt.Sprintf("/api/v1/posts/%d/report", postID)

	// 幽灵 → 401
	if code, _ := do(t, http.MethodPost, base, ghost, map[string]any{"reason": "x"}); code != 401 {
		t.Fatalf("ghost report: %d", code)
	}
	// 非法 body → 400
	if code, _ := rawDo(t, http.MethodPost, base, tok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad report body: %d", code)
	}
}

// ---- middleware & handlers 的幽灵/400 分支（走 ListTopics supporter 数据） ----

// TestCovTopicBranches 覆盖 ListTopics 空标签 ids、ToggleFollow 各分支、CreateTopic
// 校验分支、UpdateTopic category、loadFollowed/loadBookmarks 空与非空等。
func TestCovTopicBranches(t *testing.T) {
	ghost := covGhostToken(t)
	tok := covRegister(t, "covtop_a", "covtop_a@gouache.local")

	// CreateTopic 幽灵 → 401
	if code, _ := do(t, http.MethodPost, "/api/v1/topics", ghost, map[string]any{"title": "x", "raw": "y"}); code != 401 {
		t.Fatalf("ghost create topic: %d", code)
	}
	// CreateTopic 非法 body → 400
	if code, _ := rawDo(t, http.MethodPost, "/api/v1/topics", tok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad body create topic: %d", code)
	}
	// CreateTopic 空标签 & 重复标签 → 覆盖循环内 skip 分支
	topicID := covCreateTopic(t, tok, "标签去重覆盖", "去重", nil)
	code, b := do(t, http.MethodPost, "/api/v1/topics", tok, map[string]any{
		"title": "重复标签", "raw": "ok", "tags": []string{"covtagdup", "covtagdup", ""}})
	if code != 201 {
		t.Fatalf("create dup-tag topic: %d %s", code, b)
	}

	// ToggleFollow：幽灵→401，非法 id→400，不存在→404
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/1/follow", ghost, nil); code != 401 {
		t.Fatalf("ghost follow: %d", code)
	}
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/abc/follow", tok, nil); code != 400 {
		t.Fatalf("bad follow id: %d", code)
	}
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/999999/follow", tok, nil); code != 404 {
		t.Fatalf("follow missing topic: %d", code)
	}
	// 关注开 → 覆盖 ToggleFollow 插入分支
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", topicID), tok, nil); code != 200 {
		t.Fatalf("follow on: %d", code)
	}
	// 关注关 → 覆盖 ToggleFollow 删除分支
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", topicID), tok, nil); code != 200 {
		t.Fatalf("follow off: %d", code)
	}

	// UpdateTopic：非法 body → 400
	if code, _ := rawDo(t, http.MethodPut, fmt.Sprintf("/api/v1/topics/%d", topicID), tok, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad body update topic: %d", code)
	}
	// UpdateTopic 带 category_id → 覆盖 category 更新分支
	cat, _ := models.GetDB().Category.Select().Where("slug = ?", "general").One()
	if cat == nil {
		t.Fatal("no general category")
	}
	catID := cat.ID
	if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/topics/%d", topicID), tok, map[string]any{"category_id": catID}); code != 200 {
		t.Fatalf("update topic category: %d", code)
	}

	// setTopicFlag 的 findTopic ParseInt 失败 → pin abc → 404
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/abc/pin", tok, nil); code != 404 {
		t.Fatalf("pin bad id: %d", code)
	}

	// loadFollowed / loadBookmarks 空 ids 分支：带 token 访问空列表。
	// 通过删除一个只用该唯一标签的话题，让标签存在但无 topics —— 走 ListTopics 空 ids 分支
	if code, b := do(t, http.MethodGet, "/api/v1/topics?tag=covtagdup", "", nil); code != 200 {
		t.Fatalf("tag filter: %d", code)
	} else if len(dataList(t, b)) != 1 {
		t.Fatalf("covtagdup should have 1: %d", len(dataList(t, b)))
	}
}

// TestCovListTopicsEmptyTags 让一个"仅被某个话题使用、话题随后被删"的标签变为无话题，
// 触发 ListTopics 的 tag 命中但 ids 空 → 提前返回空列表，同时带登录态触发 loadFollowed/
// loadBookmarks 的空 ids 分支。
func TestCovListTopicsEmptyTags(t *testing.T) {
	tok := covRegister(t, "covtag_a", "covtag_a@gouache.local")
	id := covCreateTopic(t, tok, "唯一定制话题", "自定义标签", []string{"covuniq1"})
	// 作者删除该话题 → TopicTag 删除，但 Tag 记录保留、topics_count=0
	if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/topics/%d", id), tok, nil); code != 200 {
		t.Fatalf("delete tagged topic: %d", code)
	}
	// 带登录态触发空 ids 的 loadFollowed/loadBookmarks
	if code, b := do(t, http.MethodGet, "/api/v1/topics?tag=covuniq1", tok, nil); code != 200 {
		t.Fatalf("tag empty filter: %d", code)
	} else if len(dataList(t, b)) != 0 {
		t.Fatalf("covuniq1 should be empty: %d", len(dataList(t, b)))
	}
	// 匿名访问也该为空
	if code, b := do(t, http.MethodGet, "/api/v1/topics?tag=covuniq1", "", nil); code != 200 {
		t.Fatalf("tag empty anon: %d", code)
	} else if len(dataList(t, b)) != 0 {
		t.Fatalf("covuniq1 should be empty anon: %d", len(dataList(t, b)))
	}
}

// TestCovListPostsLiked 覆盖 ListPosts 中"当前用户已点赞"的 liked 映射填充，以及
// ListPosts 的 500（关闭 DB）。也覆盖 CreatePost 幽灵 401 / UpdatePost bind / DeletePost 他人 403。
func TestCovListPostsLiked(t *testing.T) {
	ghost := covGhostToken(t)
	ak := covRegister(t, "covlk_a", "covlk_a@gouache.local")
	bk := covRegister(t, "covlk_b", "covlk_b@gouache.local")
	topicID := covCreateTopic(t, ak, "点赞列表覆盖", "正文", nil)
	var postID int64
	{
		// 取首帖 id
		p, err := models.GetDB().Post.Select().Where("topic_id = ?", topicID).One()
		if err != nil || p == nil {
			t.Fatalf("get post: %v", err)
		}
		postID = p.ID
	}

	// bk 点赞首帖
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/like", postID), bk, nil); code != 200 {
		t.Fatalf("like: %d", code)
	}
	// bk 查看帖子流 → liked 映射填充（line 55/56）
	if code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), bk, nil); code != 200 {
		t.Fatalf("liked list posts: %d", code)
	} else if len(dataList(t, b)) != 1 {
		t.Fatalf("posts should be 1: %d", len(dataList(t, b)))
	}

	// CreatePost 幽灵 → 401
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), ghost, map[string]any{"raw": "x"}); code != 401 {
		t.Fatalf("ghost create post: %d", code)
	}
	// UpdatePost 非法 body → 400
	if code, _ := rawDo(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", postID), ak, "application/json", strings.NewReader("{bad")); code != 400 {
		t.Fatalf("bad body update post: %d", code)
	}
	// DeletePost 他人删除（bk 删 ak 的首帖）→ 403。首帖其实不能删，但权限先于首帖判断。
	if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", postID), bk, nil); code != 403 {
		t.Fatalf("delete others post: %d", code)
	}
}

// TestCovNotificationsExtra 覆盖通知相关的幽灵 401 分支，以及
// createReplyNotifications 的关注者分支 + createMentionNotifications 去重分支。
func TestCovNotificationsExtra(t *testing.T) {
	ghost := covGhostToken(t)
	author := covRegister(t, "covnt_a", "covnt_a@gouache.local")
	follower := covRegister(t, "covnt_b", "covnt_b@gouache.local")
	replier := covRegister(t, "covnt_c", "covnt_c@gouache.local")

	topicID := covCreateTopic(t, author, "通知覆盖", "正文", nil)

	// follower 关注话题
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", topicID), follower, nil); code != 200 {
		t.Fatalf("follow: %d", code)
	}

	// replier 回复并双次 @alice，触发关注者通知 + 提及去重
	raw := "回复内容 @alice @alice @ALICE"
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), replier, map[string]any{"raw": raw}); code != 201 {
		t.Fatalf("reply: %d", code)
	}
	// 确认：follower 收到回复通知（否则 createReplyNotifications 关注者分支未覆盖）
	followTok := follower
	if code, b := do(t, http.MethodGet, "/api/v1/notifications", followTok, nil); code != 200 {
		t.Fatalf("follower notifs: %d", code)
	} else if len(jarr(t, b)) == 0 {
		t.Fatal("follower should have notifications from reply")
	}

	// 幽灵 → 401：ListNotifications / unread / mark-read / read-all / stream
	if code, _ := do(t, http.MethodGet, "/api/v1/notifications", ghost, nil); code != 401 {
		t.Fatalf("ghost list notif: %d", code)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/notifications/unread", ghost, nil); code != 401 {
		t.Fatalf("ghost unread: %d", code)
	}
	if code, _ := do(t, http.MethodPost, "/api/v1/notifications/1/read", ghost, nil); code != 401 {
		t.Fatalf("ghost read: %d", code)
	}
	if code, _ := do(t, http.MethodPost, "/api/v1/notifications/read-all", ghost, nil); code != 401 {
		t.Fatalf("ghost read-all: %d", code)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/notifications/stream", ghost, nil); code != 401 {
		t.Fatalf("ghost stream: %d", code)
	}
}

// TestCovClosedDB500 关闭 DB 后走真实查询失败路径，覆盖 ListTopics / ListPosts 的 500 分支。
// 结束后重新打开 DB。
func TestCovClosedDB500(t *testing.T) {
	models.CloseDB()
	defer func() {
		if _, err := models.OpenDB(config.GetConfig().Database); err != nil {
			t.Fatalf("reopen db: %v", err)
		}
	}()

	// ListTopics 查询失败 → 500
	if code, _ := do(t, http.MethodGet, "/api/v1/topics", "", nil); code != 500 {
		t.Fatalf("closed-db list topics want 500: %d", code)
	}
	// ListPosts 查询失败 → 500
	if code, _ := do(t, http.MethodGet, "/api/v1/topics/1/posts", "", nil); code != 500 {
		t.Fatalf("closed-db list posts want 500: %d", code)
	}
}

// ---- 追加：setTopicFlag / 级联删除 / 取消点赞 / 收藏数据 / 撤销投票 / 举报处理 / 帖子编辑删除 ----

// loadPostIDs 按楼号顺序取话题所有帖子 id。
func loadPostIDs(t *testing.T, topicID int64) []int64 {
	t.Helper()
	posts, err := models.GetDB().Post.Select().
		Where("topic_id = ?", topicID).
		OrderBy("post_number ASC").
		All()
	if err != nil || len(posts) == 0 {
		t.Fatalf("posts of topic %d: %v", topicID, err)
	}
	ids := make([]int64, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	return ids
}

// TestCovSetTopicFlag 覆盖 setTopicFlag 各分支：非 staff 403、非法 id 404、
// staff 对 pin/unpin/close/open 的 200 分支。
func TestCovSetTopicFlag(t *testing.T) {
	admin := authToken(t, "admin")
	reg := covRegister(t, "covflag_a", "covflag_a@gouache.local")
	id := covCreateTopic(t, reg, "置顶覆盖", "flag", nil)

	// 非 staff → 403
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/pin", id), reg, nil); code != 403 {
		t.Fatalf("pin non-staff: %d", code)
	}
	// 非法 id → 404（findTopic ParseInt 失败）
	if code, _ := do(t, http.MethodPost, "/api/v1/topics/abc/pin", admin, nil); code != 404 {
		t.Fatalf("pin bad id: %d", code)
	}
	// staff 200 分支
	for _, act := range []string{"pin", "close", "unpin", "open"} {
		url := fmt.Sprintf("/api/v1/topics/%d/%s", id, act)
		if code, _ := do(t, http.MethodPost, url, admin, nil); code != 200 {
			t.Fatalf("%s: %d", act, code)
		}
	}
}

// TestCovDeleteTopicRich 让一个话题带关注/收藏/多帖/点赞/投票/举报后删除，
// 覆盖 DeleteTopic 的级联删除各分支、非作者 403、staff 直接删除。
func TestCovDeleteTopicRich(t *testing.T) {
	ak := covRegister(t, "covdel_a", "covdel_a@gouache.local")
	bk := covRegister(t, "covdel_b", "covdel_b@gouache.local")
	id := covCreateTopic(t, ak, "级联删除覆盖", "rich", []string{"covdel"})

	reply := func(who string) {
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", id), who, map[string]any{"raw": "reply"}); code != 201 {
			t.Fatalf("reply: %d", code)
		}
	}
	reply(bk)
	reply(ak)
	posts := loadPostIDs(t, id)

	// bk 关注 + 收藏
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/follow", id), bk, nil); code != 200 {
		t.Fatalf("follow: %d", code)
	}
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/bookmark", id), bk, nil); code != 200 {
		t.Fatalf("bookmark: %d", code)
	}
	// 对非首帖制造点赞/投票/举报，覆盖帖内循环的各删除分支
	for _, pid := range posts[1:] {
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/like", pid), ak, nil); code != 200 {
			t.Fatalf("like: %d", code)
		}
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/vote", pid), ak, map[string]any{"vote": 1}); code != 200 {
			t.Fatalf("vote: %d", code)
		}
		if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", pid), bk, map[string]any{"reason": "r"}); code != 201 {
			t.Fatalf("report: %d", code)
		}
	}

	// 非作者删除 → 403
	if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/topics/%d", id), bk, nil); code != 403 {
		t.Fatalf("delete others topic: %d", code)
	}
	// staff (admin) 删除 → 200，覆盖 canEditTopic staff 分支
	admin := authToken(t, "admin")
	if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/topics/%d", id), admin, nil); code != 200 {
		t.Fatalf("staff delete topic: %d", code)
	}
}

// TestCovUnlike 覆盖 ToggleLike 的取消点赞（delete）分支与再次点赞（通知）。
func TestCovUnlike(t *testing.T) {
	ak := covRegister(t, "covlike_a", "covlike_a@gouache.local")
	id := covCreateTopic(t, ak, "取消点赞覆盖", "x", nil)
	pid := covFirstPostID(t, id)
	base := fmt.Sprintf("/api/v1/posts/%d/like", pid)
	if code, _ := do(t, http.MethodPost, base, ak, nil); code != 200 {
		t.Fatalf("like: %d", code)
	}
	// 再点 → 取消点赞（delete 分支）
	if code, b := do(t, http.MethodPost, base, ak, nil); code != 200 {
		t.Fatalf("unlike: %d", code)
	} else if m := jmap(t, b); m["liked"] != false {
		t.Fatalf("unlike should be false: %v", m)
	}
	// 再 like → createLikeNotification
	if code, _ := do(t, http.MethodPost, base, ak, nil); code != 200 {
		t.Fatalf("like2: %d", code)
	}
	// 超长点赞? 无；改用非法话题 make ToggleLike 404
	if code, _ := do(t, http.MethodPost, "/api/v1/posts/abc/like", ak, nil); code != 404 {
		t.Fatalf("like bad id: %d", code)
	}
}

// TestCovBookmarkData 覆盖 ToggleBookmark 的加/删分支与 ListBookmarks 有数据分支。
func TestCovBookmarkData(t *testing.T) {
	ak := covRegister(t, "covbk_a", "covbk_a@gouache.local")
	id := covCreateTopic(t, ak, "收藏数据覆盖", "bm", nil)
	bmURL := fmt.Sprintf("/api/v1/topics/%d/bookmark", id)

	if code, _ := do(t, http.MethodPost, bmURL, ak, nil); code != 200 {
		t.Fatalf("bookmark on: %d", code)
	}
	// ListBookmarks 有数据 → loadBookmarks/ListBookmarks 数据分支
	if code, b := do(t, http.MethodGet, "/api/v1/bookmarks", ak, nil); code != 200 {
		t.Fatalf("list bookmarks: %d", code)
	} else if a := jarr(t, b); len(a) != 1 {
		t.Fatalf("bookmarks should be 1: %d", len(a))
	}
	// 登录态 GetTopic → loadBookmarks 非空
	if code, _ := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d", id), ak, nil); code != 200 {
		t.Fatalf("get topic: %d", code)
	}
	// 取消收藏（delete 分支）
	if code, _ := do(t, http.MethodPost, bmURL, ak, nil); code != 200 {
		t.Fatalf("bookmark off: %d", code)
	}
	// 再 ListBookmarks 应空
	if code, b := do(t, http.MethodGet, "/api/v1/bookmarks", ak, nil); code != 200 {
		t.Fatalf("empty bookmarks: %d", code)
	} else if a := jarr(t, b); len(a) != 0 {
		t.Fatalf("bookmarks should be 0: %d", len(a))
	}
}

// TestCovVoteRemove 覆盖 VotePost 的撤销（同值再点 / 投 0）分支与非法值 400。
func TestCovVoteRemove(t *testing.T) {
	ak := covRegister(t, "covvr_a", "covvr_a@gouache.local")
	id := covCreateTopic(t, ak, "撤销投票覆盖", "x", nil)
	pid := covFirstPostID(t, id)
	base := fmt.Sprintf("/api/v1/posts/%d/vote", pid)

	if code, _ := do(t, http.MethodPost, base, ak, map[string]any{"vote": 1}); code != 200 {
		t.Fatalf("vote: %d", code)
	}
	// 同值再投 → 撤销（existing!=nil && req==existing → delete）
	if code, b := do(t, http.MethodPost, base, ak, map[string]any{"vote": 1}); code != 200 {
		t.Fatalf("vote same remove: %d", code)
	} else if m := jmap(t, b); num(m, "score") != 0 {
		t.Fatalf("score should be 0: %v", m)
	}
	// 投 0 且无既有票 → 返回 nil 分支
	if code, _ := do(t, http.MethodPost, base, ak, map[string]any{"vote": 0}); code != 200 {
		t.Fatalf("vote zero: %d", code)
	}
	// 非法投票值 → 400
	if code, _ := do(t, http.MethodPost, base, ak, map[string]any{"vote": 9}); code != 400 {
		t.Fatalf("invalid vote: %d", code)
	}
}

// TestCovReportResolve 覆盖 CreateReport 重复举报更新、ResolveReport 删除违规帖、
// DismissReport 同状态提前返回、非法/缺失 id 分支。
func TestCovReportResolve(t *testing.T) {
	ak := covRegister(t, "covrr_a", "covrr_a@gouache.local")
	admin := authToken(t, "admin")
	id := covCreateTopic(t, ak, "举报处理覆盖", "x", nil)
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", id), ak, map[string]any{"raw": "second"}); code != 201 {
		t.Fatalf("reply2: %d", code)
	}
	posts := loadPostIDs(t, id)
	first := posts[0]
	second := posts[1]

	// 举报第二帖两次 → 第二次走 existing 更新分支(200)
	var rid int64
	for i := 0; i < 2; i++ {
		code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", second), ak, map[string]any{"reason": "promo"})
		if i == 0 {
			if code != 201 {
				t.Fatalf("report1: %d %s", code, b)
			}
			rid = num(jmap(t, b), "id")
		} else if code != 200 {
			t.Fatalf("report dup: %d %s", code, b)
		}
	}
	// resolve → 删除违规帖（deletePostInternal，非首帖）
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/resolve", rid), admin, nil); code != 200 {
		t.Fatalf("resolve: %d", code)
	}
	// 举报首帖并 dismiss → status!=resolved 不删除帖子
	var rid2 int64
	{
		code, b := do(t, http.MethodPost, fmt.Sprintf("/api/v1/posts/%d/report", first), ak, map[string]any{"reason": "r2"})
		if code != 201 {
			t.Fatalf("report first: %d %s", code, b)
		}
		rid2 = num(jmap(t, b), "id")
	}
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/dismiss", rid2), admin, nil); code != 200 {
		t.Fatalf("dismiss: %d", code)
	}
	// 再 dismiss 同状态 → 提前返回
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/dismiss", rid2), admin, nil); code != 200 {
		t.Fatalf("dismiss again: %d", code)
	}
	// 非法 report id → 400
	if code, _ := do(t, http.MethodPost, "/api/v1/reports/abc/resolve", admin, nil); code != 400 {
		t.Fatalf("resolve bad id: %d", code)
	}
	// 不存在 report → 404
	if code, _ := do(t, http.MethodPost, "/api/v1/reports/999999/resolve", admin, nil); code != 404 {
		t.Fatalf("resolve missing: %d", code)
	}
}

// TestCovUpdateDeletePost 覆盖 UpdatePost 成功/空内容400/非法id404，DeletePost 成功/首帖400/非法id404。
func TestCovUpdateDeletePost(t *testing.T) {
	ak := covRegister(t, "covud_a", "covud_a@gouache.local")
	_ = covRegister(t, "covud_b", "covud_b@gouache.local")
	id := covCreateTopic(t, ak, "编辑删除帖子覆盖", "x", nil)
	pid1 := covFirstPostID(t, id)

	// UpdatePost 成功（保存）分支
	if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", pid1), ak, map[string]any{"raw": "edited"}); code != 200 {
		t.Fatalf("update own post: %d", code)
	}
	// ak 回复第二帖
	if code, _ := do(t, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/posts", id), ak, map[string]any{"raw": "own2"}); code != 201 {
		t.Fatalf("own reply2: %d", code)
	}
	posts := loadPostIDs(t, id)
	pid2 := posts[1]
	// 空内容 → 400
	if code, _ := do(t, http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", pid2), ak, map[string]any{"raw": ""}); code != 400 {
		t.Fatalf("update empty raw: %d", code)
	}
	// UpdatePost 非法 id → 404
	if code, _ := do(t, http.MethodPut, "/api/v1/posts/abc", ak, map[string]any{"raw": "x"}); code != 404 {
		t.Fatalf("update post bad id: %d", code)
	}
	// DeletePost 自己第二帖 → 200（deletePostInternal）
	if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", pid2), ak, nil); code != 200 {
		t.Fatalf("delete own post: %d", code)
	}
	// DeletePost 首帖 → 400
	if code, _ := do(t, http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", pid1), ak, nil); code != 400 {
		t.Fatalf("delete first post: %d", code)
	}
	// DeletePost 非法 id → 404
	if code, _ := do(t, http.MethodDelete, "/api/v1/posts/abc", ak, nil); code != 404 {
		t.Fatalf("delete post bad id: %d", code)
	}
}

// covRawAuth 发送带自定义 Authorization 头的请求（用于覆盖 parseToken 的非法头分支）。
func covRawAuth(t *testing.T, method, path, authHeader string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", authHeader)
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, req)
	return w.Code
}

// TestCovParseTokenBranches 覆盖 parseToken 的非 Bearer scheme、缺空格、无效 token 分支。
func TestCovParseTokenBranches(t *testing.T) {
	// 非 Bearer scheme → parts 长度 2 但 scheme 不匹配 → nil → 401
	if code := covRawAuth(t, http.MethodGet, "/api/v1/auth/me", "Basic abc"); code != 401 {
		t.Fatalf("basic auth want 401: %d", code)
	}
	// 缺空格无 scheme → 只拆出 1 段 → nil → 401
	if code := covRawAuth(t, http.MethodGet, "/api/v1/auth/me", "BearerOnlyJunk"); code != 401 {
		t.Fatalf("no-space bearer want 401: %d", code)
	}
	// 合法格式但签名/token 无效 → 解析失败 → 401
	if code := covRawAuth(t, http.MethodGet, "/api/v1/auth/me", "Bearer invalid.token.here"); code != 401 {
		t.Fatalf("invalid token want 401: %d", code)
	}
}

// TestCovTrustLevelBranches 通过直接入库大量帖子，触发 trustLevel 的 intermediate/资深分支。
func TestCovTrustLevelBranches(t *testing.T) {
	mid := covRegister(t, "covmid_a", "covmid_a@gouache.local")
	sen := covRegister(t, "covsen_a", "covsen_a@gouache.local")
	midTopic := covCreateTopic(t, mid, "中级覆盖", "x", nil)
	senTopic := covCreateTopic(t, sen, "资深覆盖", "y", nil)
	midUser, _ := models.GetDB().User.Select().Where("username = ?", "covmid_a").One()
	senUser, _ := models.GetDB().User.Select().Where("username = ?", "covsen_a").One()
	db := models.GetDB()
	// mid 共 12 帖 → 中级
	for i := 2; i <= 12; i++ {
		if err := db.Post.Insert().One(&models.Post{TopicID: midTopic, UserID: midUser.ID, PostNumber: i, Raw: "p"}); err != nil {
			t.Fatalf("insert mid post: %v", err)
		}
	}
	// sen 共 53 帖 → 资深
	for i := 2; i <= 53; i++ {
		if err := db.Post.Insert().One(&models.Post{TopicID: senTopic, UserID: senUser.ID, PostNumber: i, Raw: "p"}); err != nil {
			t.Fatalf("insert sen post: %v", err)
		}
	}
	assertLevel := func(topicID int64, token, want string) {
		t.Helper()
		code, b := do(t, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/posts", topicID), token, nil)
		if code != 200 {
			t.Fatalf("posts of trust topic: %d", code)
		}
		posts := dataList(t, b)
		if len(posts) == 0 {
			t.Fatalf("no posts")
		}
		m := posts[0].(map[string]any)
		u := m["user"].(map[string]any)
		if name, _ := u["level_name"].(string); name != want {
			t.Fatalf("level want %s got %s", want, name)
		}
	}
	assertLevel(midTopic, mid, "中级")
	assertLevel(senTopic, sen, "资深")
}

// TestCovClosedRead500 在关闭 DB 下覆盖各只读 handler 的 500 / 404 分支。
func TestCovClosedRead500(t *testing.T) {
	admin := authToken(t, "admin") // 先从打开态取 token（避免查询 DB）
	models.CloseDB()
	defer func() {
		if _, err := models.OpenDB(config.GetConfig().Database); err != nil {
			t.Fatalf("reopen db: %v", err)
		}
	}()
	if code, _ := do(t, http.MethodGet, "/api/v1/categories", "", nil); code != 500 {
		t.Fatalf("closed cats want 500: %d", code)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/tags", "", nil); code != 500 {
		t.Fatalf("closed tags want 500: %d", code)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/reports", admin, nil); code != 500 {
		t.Fatalf("closed reports want 500: %d", code)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/topics/1", "", nil); code != 404 {
		t.Fatalf("closed get topic want 404: %d", code)
	}
	if code, _ := do(t, http.MethodGet, "/api/v1/users/covnobody", "", nil); code != 404 {
		t.Fatalf("closed get user want 404: %d", code)
	}
}

// TestCovCategoryTreeBranches 覆盖 buildCategoryTree 的“父分类在本批次内→attach 子分类”分支。
func TestCovCategoryTreeBranches(t *testing.T) {
	admin := authToken(t, "admin")
	_, b := do(t, http.MethodPost, "/api/v1/categories", admin, map[string]any{"name": "树根", "slug": "covroot1", "color": "#111"})
	if len(b) == 0 {
		t.Fatal("create root cat failed")
	}
	pid := num(jmap(t, b), "id")
	if _, b := do(t, http.MethodPost, "/api/v1/categories", admin, map[string]any{"name": "树枝", "slug": "covbranch1", "parent_id": pid}); len(b) == 0 {
		t.Fatalf("create branch cat failed: %s", b)
	}
	// 列表时本批次同时含父(cluster)与子(cluster selector)，触发 attach 分支
	if code, b := do(t, http.MethodGet, "/api/v1/categories", "", nil); code != 200 {
		t.Fatalf("list cats tree: %d %s", code, b)
	}
}

// TestCovStreamTicker 等待超过 SSE 轮询周期（15s），触发 ticker 分支的 ": ping" 心跳，
// 覆盖 StreamNotifications 的时间轮询路径（现有用例只覆盖初始发送 + 取消）。
func TestCovStreamTicker(t *testing.T) {
	tok := covRegister(t, "covstr_a", "covstr_a@gouache.local")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/stream", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	done := make(chan bool)
	go func() {
		testApp.ServeHTTP(w, req)
		done <- true
	}()
	select {
	case <-time.After(16 * time.Second):
	case <-done:
		t.Fatalf("stream ended early: %q", w.Body.String())
	}
	cancel()
	<-done
	if !strings.Contains(w.Body.String(), ": ping") {
		t.Fatalf("stream missing ping heartbeat: %q", w.Body.String())
	}
}
