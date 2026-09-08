package models

import (
	"fmt"

	"github.com/azhai/goent"
	"github.com/azhai/goent/drivers"
)

var db *Database

// Database 组合所有 schema 与 *goent.DB。
type Database struct {
	ForumSchema `goe:"public"`
	*goent.DB
}

func GetDB() *Database {
	return db
}

func CloseDB() {
	if db != nil {
		_ = goent.Close(db)
	}
}

func OpenDB(cfg drivers.DatabaseConfig) (*Database, error) {
	drv, err := drivers.Connect(cfg)
	if err != nil {
		return nil, err
	}
	db, err = goent.Open[Database](drv)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	if err = goent.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	// 全文检索索引（FTS5 + 触发器，幂等）
	if err = EnsureSearchIndex(db); err != nil {
		return nil, fmt.Errorf("failed to build search index: %w", err)
	}
	return db, nil
}
