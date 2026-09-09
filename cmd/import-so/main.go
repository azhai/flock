// Command import-so 从 StackOverflow Go 问答 CSV（SO-go-48000.csv）导入论坛数据。
//
// 每行 CSV 是一条“问题 + 被采纳答案”，本工具将其映射到论坛数据模型：
//   - 问题  → Topic + 首帖 Post(PostNumber=1)
//   - 答案  → 次帖 Post(PostNumber=2)
//   - 作者  → User（按用户名去重并缓存，避免重复创建）
//   - 时间  → Topic.CreatedAt / Post.CreatedAt（保留源时间戳）
//
// 运行前会先清空所有帖子相关表（Post/Topic/Tag/TopicTag/点赞/收藏/通知/举报等），
// 保留 User（admin/alice/bob）与 Category 结构数据。
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/azhai/flock/config"
	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/models"
	"github.com/azhai/gobus/environ"
	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
)

// toMarkdown 把 StackOverflow 的 HTML 正文清理成 Markdown，失败时回退原文。
func toMarkdown(htmlBody string) string {
	if htmlBody == "" {
		return ""
	}
	if md, err := htmltomarkdown.ConvertString(htmlBody); err == nil {
		return md
	}
	return htmlBody
}

// 源时间戳为无时区格式，按 UTC 解析（StackOverflow 时间戳均为 UTC）。
const timeLayout = "2006-01-02 15:04:05"

// row 对应 CSV 中的一行（见文件头：Id,Title,QuestionBody,QuestionScore,QuestionTime,
// QuestionAuthor,AnswerBody,AnswerScore,AnswerTime,AnswerAuthor）。
type row struct {
	title     string
	qBody     string
	qScore    int
	qTime     time.Time
	qTimeOK   bool
	qAuthor   string
	aBody     string
	aScore    int
	aTime     time.Time
	aTimeOK   bool
	aAuthor   string
	hasAnswer bool
}

// stats 汇总导入结果，便于日志与核对。
type stats struct {
	topics  int
	posts   int
	users   int
	skipped int
}

func main() {
	csvPath := flag.String("csv", "SO-go-48000.csv", "StackOverflow Go 问答 CSV 路径")
	batch := flag.Int("batch", 100, "每批提交的事务行数")
	flag.Parse()

	env := environ.NewEmptyEnviron()
	if _, err := os.Stat(".env"); err == nil {
		env = environ.NewEnv()
	}
	cfg := config.Load(env)
	if _, err := models.OpenDB(cfg.Database); err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer models.CloseDB()

	clearPosts()
	st, err := importCSV(*csvPath, *batch)
	if err != nil {
		log.Fatalf("导入失败: %v", err)
	}
	if err := models.RebuildSearchIndex(models.GetDB()); err != nil {
		log.Fatalf("重建全文索引失败: %v", err)
	}
	log.Printf("导入完成: topics=%d posts=%d users=%d skipped=%d",
		st.topics, st.posts, st.users, st.skipped)
}

// clearPosts 清空所有帖子及其关联表，保留 User 与 Category。
func clearPosts() {
	db := models.GetDB()
	tables := []func() error{
		func() error { return db.PostVote.Delete().Exec() },
		func() error { return db.PostLike.Delete().Exec() },
		func() error { return db.TopicFollower.Delete().Exec() },
		func() error { return db.Bookmark.Delete().Exec() },
		func() error { return db.Notification.Delete().Exec() },
		func() error { return db.Report.Delete().Exec() },
		func() error { return db.TopicTag.Delete().Exec() },
		func() error { return db.Tag.Delete().Exec() },
		func() error { return db.Post.Delete().Exec() },
		func() error { return db.Topic.Delete().Exec() },
	}
	for _, del := range tables {
		if err := del(); err != nil {
			log.Fatalf("清空表失败: %v", err)
		}
	}
	log.Println("已清空帖子相关表")
}

// importCSV 流式读取 CSV 并分批事务写入。
func importCSV(path string, batch int) (*stats, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db := models.GetDB()
	st := &stats{}
	users := map[string]int64{}

	r := csv.NewReader(f)
	r.FieldsPerRecord = 10
	r.ReuseRecord = true

	// 读取表头（数据从第二行开始）。
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("读取表头失败: %w", err)
	}

	// 分批事务：每 batch 行提交一次，兼顾回滚粒度与性能。
	collect := []func(tx model.Transaction) error{}
	commit := func() error {
		if len(collect) == 0 {
			return nil
		}
		err := db.BeginTransaction(func(tx model.Transaction) error {
			for _, fn := range collect {
				if err := fn(tx); err != nil {
					return err
				}
			}
			return nil
		})
		collect = collect[:0]
		return err
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			st.skipped++
			log.Printf("解析行失败(skipped): %v", err)
			continue
		}

		rv := parseRow(rec)
		collect = append(collect, buildInsert(rv, users, st))

		if len(collect) >= batch {
			if err := commit(); err != nil {
				return nil, err
			}
			log.Printf("已导入 topics=%d posts=%d users=%d", st.topics, st.posts, st.users)
		}
	}
	return st, commit()
}

// parseRow 将原始 CSV 字段解析为结构化行；时间解析失败时保留零值并置对应 OK 标志为 false。
func parseRow(rec []string) row {
	rv := row{
		title:   strings.TrimSpace(rec[1]),
		qBody:   rec[2],
		qAuthor: strings.TrimSpace(rec[5]),
		aBody:   rec[6],
		aAuthor: strings.TrimSpace(rec[9]),
	}
	rv.qScore = atoi(rec[3])
	rv.aScore = atoi(rec[7])
	if t, err := time.Parse(timeLayout, strings.TrimSpace(rec[4])); err == nil {
		rv.qTime, rv.qTimeOK = t, true
	}
	if t, err := time.Parse(timeLayout, strings.TrimSpace(rec[8])); err == nil {
		rv.aTime, rv.aTimeOK = t, true
	}
	rv.hasAnswer = strings.TrimSpace(rv.aBody) != ""
	if rv.qAuthor == "" {
		rv.qAuthor = "anonymous"
	}
	return rv
}

// buildInsert 构造单个记录的所有写操作，统一放入当前事务。
func buildInsert(rv row, users map[string]int64, st *stats) func(tx model.Transaction) error {
	return func(tx model.Transaction) error {
		db := models.GetDB()
		qUserID, err := ensureUser(tx, rv.qAuthor, users, st)
		if err != nil {
			return err
		}
		created := rv.qTime

		topic := &models.Topic{
			Title:      rv.title,
			Slug:       helpers.Slugify(rv.title),
			CategoryID: nil,
			UserID:     qUserID,
			CreatedAt:  created,
			UpdatedAt:  created,
			LikeCount:  rv.qScore,
			PostsCount: 1,
		}
		if err := db.Topic.Insert().OnTransaction(tx).One(topic); err != nil {
			return err
		}
		st.topics++
		posts := int64(1)
		lastID := insertPost(tx, topic.ID, qUserID, 1, toMarkdown(rv.qBody), rv.qScore, created, st)
		lastPosted := created

		if rv.hasAnswer {
			aUserID := ensureSameUser(qUserID, rv.qAuthor, rv.aAuthor, tx, users, st)
			aTime := created
			if rv.aTimeOK {
				aTime = rv.aTime
			}
			lastID = insertPost(tx, topic.ID, aUserID, 2, toMarkdown(rv.aBody), rv.aScore, aTime, st)
			posts = 2
			lastPosted = aTime
		}

		return db.Topic.Update().
			Set(
				goent.Pair{Key: "posts_count", Value: posts},
				goent.Pair{Key: "reply_count", Value: posts - 1},
				goent.Pair{Key: "last_post_id", Value: lastID},
				goent.Pair{Key: "last_posted_at", Value: lastPosted},
				goent.Pair{Key: "updated_at", Value: lastPosted},
			).
			Filter(goent.Equals(db.Topic.Field("id"), topic.ID)).
			OnTransaction(tx).
			Exec()
	}
}

// insertPost 插入一个楼层帖子并更新计数。
func insertPost(tx model.Transaction, topicID, userID int64, num int, body string, score int, at time.Time, st *stats) int64 {
	db := models.GetDB()
	p := &models.Post{
		TopicID:    topicID,
		UserID:     userID,
		PostNumber: num,
		Raw:        body,
		LikeCount:  score,
		CreatedAt:  at,
		UpdatedAt:  at,
	}
	if err := db.Post.Insert().OnTransaction(tx).One(p); err != nil {
		// 事务内失败会由外层返回，这里仅递增失败不中断（外层会回滚整批）。
		log.Printf("插入 post 失败 (topic=%d post=%d): %v", topicID, num, err)
		return 0
	}
	st.posts++
	return p.ID
}

// ensureSameUser 当作者字段非空且与问题作者不同时，创建答案作者并返回其 ID。
func ensureSameUser(qID int64, qAuthor, aAuthor string, tx model.Transaction, users map[string]int64, st *stats) int64 {
	if aAuthor == "" || aAuthor == qAuthor {
		return qID
	}
	id, err := ensureUser(tx, aAuthor, users, st)
	if err != nil {
		log.Printf("创建答案作者失败 (%s): %v", aAuthor, err)
		return qID
	}
	return id
}

// ensureUser 按用户名查找或创建作者；缓存于内存避免重复查询与唯一键冲突。
func ensureUser(tx model.Transaction, username string, users map[string]int64, st *stats) (int64, error) {
	if id, ok := users[username]; ok {
		return id, nil
	}
	db := models.GetDB()
	u, err := db.User.Select().
		Where("username = ?", username).
		OnTransaction(tx).One()
	if err == nil && u != nil {
		users[username] = u.ID
		return u.ID, nil
	}
	// 导入用户不参与登录，故不生成 bcrypt（48000 量级作者逐个 bcrypt 极慢）。
	// Email 取用户名字节的十六进制，保证不同用户名 → 不同 Email（注入式，避免唯一键冲突）。
	u = &models.User{
		Username: username,
		Email:    fmt.Sprintf("so_%x@so.local", []byte(username)),
		Password: fmt.Sprintf("so-import-%d", time.Now().UnixNano()),
		Role:     models.RoleUser,
		IsActive: true,
	}
	if err := db.User.Insert().OnTransaction(tx).One(u); err != nil {
		return 0, err
	}
	st.users++
	users[username] = u.ID
	return u.ID, nil
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
