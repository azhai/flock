package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/azhai/gobus/environ"
	"github.com/azhai/gouache/config"
	"github.com/azhai/gouache/models"
	"github.com/labstack/echo/v5"
)

func setupConfig(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("TOKEN_EXPIRY", "72")
	config.Load(environ.NewEmptyEnviron())
}

func tokenFor(t *testing.T, user *models.User) string {
	setupConfig(t)
	tok, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return tok
}

func newCtx(token string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func runMW(t *testing.T, mw echo.MiddlewareFunc, authHeader string) (*httptest.ResponseRecorder, int) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var nextCalled bool
	h := func(cc *echo.Context) error {
		nextCalled = true
		return cc.JSON(200, "ok")
	}
	if err := mw(h)(c); err != nil {
		t.Fatalf("middleware err: %v", err)
	}
	if nextCalled {
		return rec, 200
	}
	return rec, rec.Code
}

func TestGenerateAndParseToken(t *testing.T) {
	u := &models.User{ID: 7, Username: "bob", Email: "b@x.io", Role: models.RoleUser}
	tok := tokenFor(t, u)
	c, _ := newCtx(tok)
	claims := parseToken(c)
	if claims == nil || claims.UserID != 7 || claims.Username != "bob" {
		t.Fatalf("parseToken bad: %+v", claims)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	setupConfig(t)
	u := &models.User{ID: 1, Username: "a", Email: "a@x.io"}
	tok := tokenFor(t, u)

	// 无 header
	c1, _ := newCtx("")
	if parseToken(c1) != nil {
		t.Fatal("empty header should be nil")
	}
	// header 无 Bearer（错误 scheme）
	c2, _ := newCtx(tok)
	c2.Request().Header.Set("Authorization", "Basic abc")
	if parseToken(c2) != nil {
		t.Fatal("wrong scheme should be nil")
	}
	// 单段
	c3, _ := newCtx(tok)
	c3.Request().Header.Set("Authorization", "BearerOnly")
	if parseToken(c3) != nil {
		t.Fatal("single-part should be nil")
	}
	// 无效 token
	c4, _ := newCtx(tok)
	c4.Request().Header.Set("Authorization", "Bearer not.a.jwt")
	if parseToken(c4) != nil {
		t.Fatal("invalid token should be nil")
	}
}

func TestAuthMiddleware(t *testing.T) {
	setupConfig(t)
	u := &models.User{ID: 1, Username: "a", Email: "a@x.io"}
	tok := tokenFor(t, u)

	// 有效
	if rec, code := runMW(t, AuthMiddleware(), "Bearer "+tok); code != 200 || rec.Code != 200 {
		t.Fatalf("valid token: code=%d", code)
	}
	// 无 token → 401
	if rec, _ := runMW(t, AuthMiddleware(), ""); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	// 非法 token → 401
	if rec, _ := runMW(t, AuthMiddleware(), "Bearer bogus"); rec.Code != 401 {
		t.Fatalf("bad token: %d", rec.Code)
	}
}

func TestOptionalAuth(t *testing.T) {
	setupConfig(t)
	u := &models.User{ID: 5, Username: "c", Email: "c@x.io", Role: models.RoleUser}
	tok := tokenFor(t, u)
	// 无 token 也能通过
	if rec, _ := runMW(t, OptionalAuth(), ""); rec.Code != 200 {
		t.Fatalf("optional no token: %d", rec.Code)
	}
	// 有 token 通过并注入 claims
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(cc *echo.Context) error {
		if GetCurrentUserID(cc) != 5 || GetCurrentUsername(cc) != "c" || GetCurrentUserRole(cc) != models.RoleUser {
			return cc.JSON(500, "claims missing")
		}
		return cc.JSON(200, "ok")
	}
	_ = OptionalAuth()(h)(c)
	if rec.Code != 200 {
		t.Fatalf("optional auth claims: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func runRoleMW(t *testing.T, roleMW echo.MiddlewareFunc, authHeader string) *httptest.ResponseRecorder {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(cc *echo.Context) error { return cc.JSON(200, "ok") }
	// 与真实路由一致：AuthMiddleware 先注入 claims，再应用角色中间件
	stack := AuthMiddleware()(roleMW(h))
	if err := stack(c); err != nil {
		t.Fatalf("middleware chain err: %v", err)
	}
	return rec
}

func TestAdminStaffOnly(t *testing.T) {
	setupConfig(t)
	adm := &models.User{ID: 1, Username: "adm", Email: "adm@x.io", Role: models.RoleAdmin}
	mod := &models.User{ID: 2, Username: "mod", Email: "mod@x.io", Role: models.RoleModerator}
	usr := &models.User{ID: 3, Username: "usr", Email: "usr@x.io", Role: models.RoleUser}
	admTok := tokenFor(t, adm)
	modTok := tokenFor(t, mod)
	usrTok := tokenFor(t, usr)

	// AdminOnly
	if rec := runRoleMW(t, AdminOnly(), "Bearer "+admTok); rec.Code != 200 {
		t.Fatalf("admin adminonly: %d %s", rec.Code, rec.Body.String())
	}
	if rec := runRoleMW(t, AdminOnly(), "Bearer "+usrTok); rec.Code != 403 {
		t.Fatalf("user adminonly: %d", rec.Code)
	}
	if rec := runRoleMW(t, AdminOnly(), "Bearer "+modTok); rec.Code != 403 {
		t.Fatalf("moderator adminonly should 403: %d", rec.Code)
	}

	// StaffOnly
	if rec := runRoleMW(t, StaffOnly(), "Bearer "+admTok); rec.Code != 200 {
		t.Fatalf("admin staffonly: %d", rec.Code)
	}
	if rec := runRoleMW(t, StaffOnly(), "Bearer "+modTok); rec.Code != 200 {
		t.Fatalf("moderator staffonly: %d", rec.Code)
	}
	if rec := runRoleMW(t, StaffOnly(), "Bearer "+usrTok); rec.Code != 403 {
		t.Fatalf("user staffonly: %d", rec.Code)
	}
}

func TestGetCurrentUserHelpers(t *testing.T) {
	c, _ := newCtx("")
	if GetCurrentUserID(c) != 0 || GetCurrentUsername(c) != "" || GetCurrentUserRole(c) != "" {
		t.Fatal("unset claims defaults wrong")
	}
	if IsStaff(c) {
		t.Fatal("unset claims should not be staff")
	}
	setClaims(c, &Claims{UserID: 2, Username: "w", Email: "w@x.io", Role: models.RoleAdmin})
	if GetCurrentUserID(c) != 2 || GetCurrentUsername(c) != "w" || GetCurrentUserRole(c) != models.RoleAdmin {
		t.Fatal("claims not applied")
	}
	if !IsStaff(c) {
		t.Fatal("admin should be staff")
	}
}

// GetCurrentUser 在未设置 user_id 时应直接返回 (nil, nil)，且不触发任何数据库访问。
func TestGetCurrentUserZeroID(t *testing.T) {
	c, _ := newCtx("")
	u, err := GetCurrentUser(c)
	if u != nil || err != nil {
		t.Fatalf("GetCurrentUser with zero id should be (nil,nil), got %v,%v", u, err)
	}
}
