package config

import (
	"testing"

	"github.com/azhai/gobus/environ"
)

// reset 清空全局缓存以便多次测试 Load。
func reset() { cfg = nil }

func TestLoadDefaults(t *testing.T) {
	reset()
	c := Load(nil)
	if c.Server.Port != 8090 || c.Server.Mode != "debug" {
		t.Errorf("server default wrong: %+v", c.Server)
	}
	if c.Server.SiteTitle != "gouache" {
		t.Errorf("site title default wrong: %q", c.Server.SiteTitle)
	}
	if c.Database.Type != "sqlite" {
		t.Errorf("db type default wrong: %q", c.Database.Type)
	}
	if c.Auth.TokenExpiry != 72 {
		t.Errorf("token expiry default wrong: %d", c.Auth.TokenExpiry)
	}
	if c.Admin.Username != "admin" || c.Admin.Email == "" {
		t.Errorf("admin default wrong: %+v", c.Admin)
	}
	if c.Database.DSN != "gouache.db" {
		t.Errorf("db dsn default wrong: %q", c.Database.DSN)
	}
}

func TestLoadWithEnv(t *testing.T) {
	reset()
	t.Setenv("APP_MODE", "release")
	t.Setenv("SERVER_PORT", "9999")
	t.Setenv("SITE_TITLE", "my-forum")
	t.Setenv("DB_DSN", "/tmp/x.db")
	t.Setenv("DB_TYPE", "sqlite")
	t.Setenv("JWT_SECRET", "s3cret")
	t.Setenv("TOKEN_EXPIRY", "12")
	t.Setenv("ADMIN_USERNAME", "root")
	t.Setenv("ADMIN_PASSWORD", "pw1234")
	t.Setenv("ADMIN_EMAIL", "r@x.io")
	t.Setenv("BASE_URL", "http://example.com")
	t.Setenv("SITE_DESCRIPTION", "desc")
	c := Load(environ.NewEmptyEnviron())
	if c.Server.Port != 9999 || c.Server.Mode != "release" {
		t.Errorf("server custom wrong: %+v", c.Server)
	}
	if c.Server.SiteTitle != "my-forum" || c.Server.SiteDescription != "desc" {
		t.Errorf("title/desc custom wrong: %+v", c.Server)
	}
	if c.Database.DSN != "/tmp/x.db" {
		t.Errorf("db dsn custom wrong: %q", c.Database.DSN)
	}
	if c.Auth.JWTSecret != "s3cret" || c.Auth.TokenExpiry != 12 {
		t.Errorf("auth custom wrong: %+v", c.Auth)
	}
	if c.Admin.Username != "root" || c.Admin.Password != "pw1234" || c.Admin.Email != "r@x.io" {
		t.Errorf("admin custom wrong: %+v", c.Admin)
	}
}

func TestLoadCaching(t *testing.T) {
	reset()
	c1 := Load(environ.NewEmptyEnviron())
	c2 := Load(environ.NewEmptyEnviron())
	if c1 != c2 {
		t.Errorf("Load should return cached cfg")
	}
}

func TestGetters(t *testing.T) {
	reset()
	Load(environ.NewEmptyEnviron())
	// 用 env 覆盖便于断言 getter 返回
	t.Setenv("JWT_SECRET", "k")
	t.Setenv("TOKEN_EXPIRY", "48")
	t.Setenv("SITE_TITLE", "s")
	reset()
	c := Load(environ.NewEmptyEnviron())
	if GetJWTSecret() != c.Auth.JWTSecret {
		t.Errorf("GetJWTSecret mismatch")
	}
	if GetTokenExpiry() != c.Auth.TokenExpiry {
		t.Errorf("GetTokenExpiry mismatch")
	}
	if GetSiteTitle() != c.Server.SiteTitle {
		t.Errorf("GetSiteTitle mismatch")
	}
	port, url := GetServerInfo()
	if port != c.Server.Port || url != c.Server.BaseURL {
		t.Errorf("GetServerInfo mismatch %d %s", port, url)
	}
	if APIBaseURL != "/api/v1" || APIVersion != "v1" {
		t.Errorf("consts wrong: %s %s", APIVersion, APIBaseURL)
	}
}

func TestGetConfig(t *testing.T) {
	reset()
	Load(environ.NewEmptyEnviron())
	if GetConfig() == nil || GetConfig().Database.DSN == "" {
		t.Errorf("GetConfig mismatch")
	}
}
