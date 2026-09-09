package models

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// FTS5 全文检索（trigram tokenizer）。
//
// 索引口径收敛为「话题标题 + 首帖纯文本摘要」，而不是把整条线程的 Markdown
// 正文（含大量代码块/标记）全部喂给 trigram——那会让索引数据随正文长度近似线性
// 膨胀到正文的 ~3 倍。摘要去 Markdown/HTML 标记后按 rune 截断，再由 Go 侧写入，
// 因此不再使用 SQL 触发器（SQL 内无法注册去标记函数）。
//
// 增量维护：只在标题变更 / 新建话题 / 首帖(PostNumber=1)变更时调用 ReindexTopic，
// 删除话题时调用 RemoveTopicIndex。只有首帖进索引，回复的增删改不影响检索。

const topicSearchSummaryRunes = 400 // 摘要最大字符数（按 rune）

// 曾用 SQL 触发器同步 topic_search 与 topic/post；现改为 Go 侧维护，启动时确保清除旧触发器，
// 避免旧版本残留触发器仍把整条线程正文灌入索引。
var obsoleteTriggers = []string{
	`DROP TRIGGER IF EXISTS trg_topic_search_ins`,
	`DROP TRIGGER IF EXISTS trg_topic_search_title`,
	`DROP TRIGGER IF EXISTS trg_topic_search_del`,
	`DROP TRIGGER IF EXISTS trg_post_search_ins`,
	`DROP TRIGGER IF EXISTS trg_post_search_upd`,
	`DROP TRIGGER IF EXISTS trg_post_search_del`,
}

// contentless 模式：不存储 title/body 的正文副本（省掉 _content 重复存储），
// 配合 content_rowid 由外部 topic 表提供展示数据；检索用 bm25() 仍可用
// （bm25 只依赖列字节长度，不依赖正文内容）。已有验证：contentless 下 bm25 排序正常。
const topicSearchDDL = `CREATE VIRTUAL TABLE IF NOT EXISTS topic_search USING fts5(title, body, content='', content_rowid='rowid', tokenize='trigram')`

// 摘要清洗用的正则（顺序敏感）：围栏代码块 → 链接/图片 → HTML 标签 → 行内代码 → 标题/引用符。
var (
	reCodeFence = regexp.MustCompile("(?s)```.*?```")
	reLink      = regexp.MustCompile(`!?\[([^\]]*)\]\([^)]*\)`)
	reHtmlTag   = regexp.MustCompile("(?s)<[^>]+>")
	reInline    = regexp.MustCompile("`[^`]*`")
	reLead      = regexp.MustCompile("(?m)^[#>\\s]+")
)

// PlainSummary 把 Markdown 正文清洗成可检索的纯文本摘要，并截断到 topicSearchSummaryRunes。
func PlainSummary(raw string) string {
	s := reCodeFence.ReplaceAllString(raw, " ")
	s = reLink.ReplaceAllString(s, "$1")
	s = reHtmlTag.ReplaceAllString(s, " ")
	s = reInline.ReplaceAllString(s, " ")
	s = reLead.ReplaceAllString(s, " ")
	plain := strings.Join(strings.Fields(s), " ") // 折叠空白与换行
	if utf8.RuneCountInString(plain) <= topicSearchSummaryRunes {
		return plain
	}
	runes := []rune(plain)
	return strings.TrimSpace(string(runes[:topicSearchSummaryRunes]))
}

// EnsureSearchIndex 确保 FTS 表就绪；当索引行数与话题数不一致时全量重建（幂等、自愈）。
// 行数一致即视为已同步，避免每次启动都重写整个索引。
func EnsureSearchIndex(db *Database) error {
	ctx := context.Background()
	for _, ddl := range obsoleteTriggers {
		if err := db.RawExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("drop obsolete search trigger: %w", err)
		}
	}
	if err := db.RawExecContext(ctx, topicSearchDDL); err != nil {
		return fmt.Errorf("create search table: %w", err)
	}

	var nSearch, nTopics int64
	_ = db.RawQueryRowContext(ctx, "SELECT count(*) FROM topic_search").Scan(&nSearch)
	_ = db.RawQueryRowContext(ctx, "SELECT count(*) FROM topic").Scan(&nTopics)
	if nSearch == nTopics && nSearch > 0 {
		return nil
	}
	return RebuildSearchIndex(db)
}

// RebuildSearchIndex 重建检索表并全量回填「标题 + 首帖摘要」。
func RebuildSearchIndex(db *Database) error {
	ctx := context.Background()
	if err := db.RawExecContext(ctx, "DROP TABLE IF EXISTS topic_search"); err != nil {
		return fmt.Errorf("drop search table: %w", err)
	}
	if err := db.RawExecContext(ctx, topicSearchDDL); err != nil {
		return fmt.Errorf("recreate search table: %w", err)
	}
	topics, err := db.Topic.Select().All()
	if err != nil {
		return fmt.Errorf("load topics for index: %w", err)
	}
	for _, t := range topics {
		var raw string
		_ = db.RawQueryRowContext(ctx,
			"SELECT IFNULL(raw,'') FROM post WHERE topic_id=? AND post_number=1", t.ID).Scan(&raw)
		if err := db.RawExecContext(ctx,
			"INSERT INTO topic_search(rowid, title, body) VALUES(?,?,?)", t.ID, t.Title, PlainSummary(raw)); err != nil {
			return fmt.Errorf("index topic %d: %w", t.ID, err)
		}
	}
	return nil
}

// ReindexTopic 重建单个话题的检索条目（新建话题、改标题、改首帖后调用）。
func ReindexTopic(topicID int64) {
	db := GetDB()
	ctx := context.Background()
	var title string
	_ = db.RawQueryRowContext(ctx, "SELECT IFNULL(title,'') FROM topic WHERE id=?", topicID).Scan(&title)
	var raw string
	_ = db.RawQueryRowContext(ctx,
		"SELECT IFNULL(raw,'') FROM post WHERE topic_id=? AND post_number=1", topicID).Scan(&raw)
	_ = db.RawExecContext(ctx, "DELETE FROM topic_search WHERE rowid=?", topicID)
	_ = db.RawExecContext(ctx,
		"INSERT INTO topic_search(rowid, title, body) VALUES(?,?,?)", topicID, title, PlainSummary(raw))
}

// RemoveTopicIndex 删除话题时移除其检索条目。
func RemoveTopicIndex(topicID int64) {
	_ = GetDB().RawExecContext(context.Background(), "DELETE FROM topic_search WHERE rowid=?", topicID)
}

// SearchTopics 按相关度返回命中的话题。匹配覆盖标题 + 首帖摘要（FTS5 trigram）；
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
		// contentless 表不能用别名做 MATCH（"no such column"），须在子查询里用全名 topic_search 检索、
		// 在外层再 JOIN topic 过滤 hidden；bm25 已在子查询内排序，外层顺序保持不变。
		var total int64
		_ = db.RawQueryRowContext(ctx,
			"SELECT count(*) FROM (SELECT rowid FROM topic_search WHERE topic_search MATCH ?) ts JOIN topic t ON t.id=ts.rowid WHERE t.hidden=0", match,
		).Scan(&total)
		if total > 0 {
			rows, err := db.RawQueryContext(ctx,
				"SELECT ts.rowid FROM (SELECT rowid FROM topic_search WHERE topic_search MATCH ? ORDER BY bm25(topic_search)) ts JOIN topic t ON t.id=ts.rowid WHERE t.hidden=0 LIMIT ? OFFSET ?",
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

	// 回退：标题模糊匹配（排除已隐藏话题）
	lq := db.Topic.Select().
		Where("title LIKE ? AND hidden = ?", "%"+q+"%", 0).
		OrderBy("pinned DESC", "created_at DESC", "id DESC")
	total, _ := lq.Count("id")
	topics, _ := lq.Skip(offset).Take(limit).All()
	for _, t := range topics {
		ids = append(ids, t.ID)
	}
	return total, ids
}
