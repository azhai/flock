package models

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// FTS5 全文检索（trigram tokenizer）。零部署（不引入独立检索服务），
// 覆盖话题标题 + 全部回复正文；中/英子串均可命中（trigram 需 >=3 字节，2 字以内查询回退 LIKE）。
// 用事务触发器维持 topic_search 与 topic/post 的增量同步。

var searchDDL = []string{
	`CREATE VIRTUAL TABLE IF NOT EXISTS topic_search USING fts5(title, body, tokenize='trigram')`,
	`CREATE TRIGGER IF NOT EXISTS trg_topic_search_ins AFTER INSERT ON topic BEGIN
		INSERT INTO topic_search(rowid, title, body)
		SELECT NEW.id, NEW.title, IFNULL((SELECT group_concat(raw, char(10)) FROM post WHERE topic_id=NEW.id), '');
	END`,
	`CREATE TRIGGER IF NOT EXISTS trg_topic_search_title AFTER UPDATE OF title ON topic BEGIN
		UPDATE topic_search SET title=NEW.title WHERE rowid=NEW.id;
	END`,
	`CREATE TRIGGER IF NOT EXISTS trg_topic_search_del AFTER DELETE ON topic BEGIN
		DELETE FROM topic_search WHERE rowid=OLD.id;
	END`,
	`CREATE TRIGGER IF NOT EXISTS trg_post_search_ins AFTER INSERT ON post BEGIN
		DELETE FROM topic_search WHERE rowid=NEW.topic_id;
		INSERT INTO topic_search(rowid, title, body)
		SELECT NEW.topic_id, t.title, IFNULL((SELECT group_concat(raw, char(10)) FROM post WHERE topic_id=NEW.topic_id), '')
		FROM topic t WHERE t.id=NEW.topic_id;
	END`,
	`CREATE TRIGGER IF NOT EXISTS trg_post_search_upd AFTER UPDATE OF raw ON post BEGIN
		DELETE FROM topic_search WHERE rowid=NEW.topic_id;
		INSERT INTO topic_search(rowid, title, body)
		SELECT NEW.topic_id, t.title, IFNULL((SELECT group_concat(raw, char(10)) FROM post WHERE topic_id=NEW.topic_id), '')
		FROM topic t WHERE t.id=NEW.topic_id;
	END`,
	`CREATE TRIGGER IF NOT EXISTS trg_post_search_del AFTER DELETE ON post BEGIN
		DELETE FROM topic_search WHERE rowid=OLD.topic_id;
		INSERT INTO topic_search(rowid, title, body)
		SELECT OLD.topic_id, t.title, IFNULL((SELECT group_concat(raw, char(10)) FROM post WHERE topic_id=OLD.topic_id), '')
		FROM topic t WHERE t.id=OLD.topic_id;
	END`,
}

// EnsureSearchIndex 建 FTS 表 + 同步触发器；仅当索引为空时全量回填一次。
func EnsureSearchIndex(db *Database) error {
	ctx := context.Background()
	for i, ddl := range searchDDL {
		if err := db.RawExecContext(ctx, ddl); err != nil &&
			!strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("search DDL[%d] %w", i, err)
		}
	}
	// goent 已修复 RawQueryRowContext：Scan 后自动关闭 rows（否则会泄漏持有读锁的连接，
	// 导致后续回填写操作 SQLITE_BUSY）。
	var n int64
	_ = db.RawQueryRowContext(ctx, "SELECT count(*) FROM topic_search").Scan(&n)
	if n > 0 {
		return nil
	}
	if err := db.RawExecContext(ctx,
		`INSERT INTO topic_search(rowid, title, body)
		 SELECT t.id, t.title, IFNULL((SELECT group_concat(p.raw, char(10)) FROM post p WHERE p.topic_id=t.id), '')
		 FROM topic t`); err != nil {
		return fmt.Errorf("search backfill: %w", err)
	}
	return nil
}

// SearchTopics 按相关度返回命中的话题。匹配覆盖标题 + 正文（FTS5 trigram）；
// 查询过短或 FTS 未命中时回退为标题 LIKE。返回 (命中总数, 本页 topic id，已按相关度排序)。
func SearchTopics(db *Database, q string, offset, limit int) (int64, []int64) {
	q = strings.TrimSpace(q)
	if q == "" {
		return 0, nil
	}
	ctx := context.Background()
	ids := make([]int64, 0, limit)

	// FTS5 trigram 需要 >=3 字节的查询词；对更短或中文双字查询直接走 LIKE 回退
	if utf8.RuneCountInString(q) >= 3 {
		match := `"` + strings.ReplaceAll(q, `"`, `""`) + `"`
		var total int64
		_ = db.RawQueryRowContext(ctx,
			"SELECT count(*) FROM topic_search WHERE topic_search MATCH ?", match,
		).Scan(&total)
		if total > 0 {
			rows, err := db.RawQueryContext(ctx,
				"SELECT rowid FROM topic_search WHERE topic_search MATCH ? ORDER BY bm25(topic_search) LIMIT ? OFFSET ?",
				match, limit, offset)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var id int64
					if rows.Scan(&id) == nil {
						ids = append(ids, id)
					}
				}
			}
			return total, ids
		}
	}

	// 回退：标题模糊匹配
	lq := db.Topic.Select().
		Where("title LIKE ?", "%"+q+"%").
		OrderBy("pinned DESC", "created_at DESC", "id DESC")
	total, _ := lq.Count("id")
	topics, _ := lq.Skip(offset).Take(limit).All()
	for _, t := range topics {
		ids = append(ids, t.ID)
	}
	return total, ids
}
