package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azhai/goent/drivers"
)

func openTempDB(t *testing.T) *Database {
	t.Helper()
	dir := t.TempDir()
	cfg := drivers.DatabaseConfig{Type: "sqlite", DSN: filepath.Join(dir, "m.db")}
	db, err := OpenDB(cfg)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(CloseDB)
	return db
}

func TestOpenAndGetDB(t *testing.T) {
	db := openTempDB(t)
	if db == nil {
		t.Fatal("db nil")
	}
	if GetDB() == nil {
		t.Fatal("GetDB nil")
	}
	// 插入与查询一个用户，验证 schema 可用
	u := &User{Username: "u1", Email: "u1@x.io", Role: RoleUser, IsActive: true}
	if err := u.SetPassword("secret1"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if err := db.User.Insert().One(u); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	got, err := db.User.Select().Where("username = ?", "u1").One()
	if err != nil || got == nil {
		t.Fatalf("select user err=%v nil=%v", err, got == nil)
	}
	if !got.CheckPassword("secret1") {
		t.Fatal("CheckPassword should pass")
	}
	if got.CheckPassword("nope") {
		t.Fatal("CheckPassword should fail")
	}
	if got.IsStaff() {
		t.Fatal("user should not be staff")
	}
}

func TestPasswordAndRoles(t *testing.T) {
	openTempDB(t)
	u := &User{Username: "a", Email: "a@x.io"}
	if err := u.SetPassword("123456"); err != nil {
		t.Fatalf("SetPassword err: %v", err)
	}
	if u.Password == "123456" {
		t.Fatal("password not hashed")
	}
	if !u.CheckPassword("123456") {
		t.Fatal("CheckPassword fail")
	}
	if u.IsStaff() {
		t.Fatal("user is not staff")
	}
	u.Role = RoleAdmin
	if !u.IsStaff() {
		t.Fatal("admin should be staff")
	}
	u.Role = RoleModerator
	if !u.IsStaff() {
		t.Fatal("moderator should be staff")
	}
}

func TestCloseDB(t *testing.T) {
	dir := t.TempDir()
	cfg := drivers.DatabaseConfig{Type: "sqlite", DSN: filepath.Join(dir, "c.db")}
	if _, err := OpenDB(cfg); err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	// 正常关闭与重复关闭都应安全（不 panic）
	CloseDB()
	CloseDB()
	// db 指针保留，但底层已关闭；再插入应失败而不是 panic
	if GetDB() == nil {
		t.Fatal("GetDB should still return the database pointer after CloseDB")
	}
	_ = os.Getpid
}
