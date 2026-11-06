package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"github.com/azhai/flock/config"
	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/models"
	"github.com/azhai/flock/routes"
	"github.com/azhai/gobus/environ"
	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

//go:embed web
var efs embed.FS

func distSub() fs.FS {
	sub, err := fs.Sub(efs, "web")
	if err != nil {
		panic(err)
	}
	return sub
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// run 承载全部启动逻辑，返回 error 以便测试覆盖（main 中任何失败统一 Fatal）。
func run() error {
	env := environ.NewEmptyEnviron()
	if _, err := os.Stat(".env"); err == nil {
		env = environ.NewEnv()
	}
	cfg := config.Load(env)
	if _, err := models.OpenDB(cfg.Database); err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer models.CloseDB()

	seedAdmin()
	seedCategories()
	seedDemo()

	app := createApp()
	port, url := config.GetServerInfo()
	log.Printf("Server starting on :%d", port)
	log.Printf("Visit %s", url)
	if err := app.Start(fmt.Sprintf(":%d", port)); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func createApp() *echo.Echo {
	dist := distSub()
	e := echo.New()
	e.Use(middleware.Recover())

	routes.SetupAPIRoutes(e)

	// 静态资源 + SPA fallback
	e.GET("/*", func(c *echo.Context) error {
		path := c.Request().URL.Path
		if strings.HasPrefix(path, "/api/") {
			return echo.NewHTTPError(404, "API endpoint not found")
		}
		name := strings.TrimPrefix(path, "/")
		if name == "" {
			name = "index.html"
		}
		if f, err := dist.Open(name); err == nil {
			if info, statErr := f.Stat(); statErr == nil && !info.IsDir() {
				_ = f.Close()
				return c.FileFS(name, dist)
			}
			_ = f.Close()
		}
		return c.FileFS("index.html", dist)
	})

	return e
}

func seedAdmin() {
	cfg := config.GetConfig()
	if cfg.Admin.Username == "" || cfg.Admin.Password == "" {
		return
	}
	db := models.GetDB()
	if u, _ := db.User.Select().Where("username = ?", cfg.Admin.Username).One(); u != nil {
		return
	}
	admin := &models.User{
		Username: cfg.Admin.Username,
		Email:    cfg.Admin.Email,
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	if err := admin.SetPassword(cfg.Admin.Password); err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}
	if err := db.User.Insert().One(admin); err != nil {
		log.Printf("Failed to seed admin: %v", err)
	} else {
		log.Printf("Seeded admin user: %s", cfg.Admin.Username)
	}
}

func seedCategories() {
	db := models.GetDB()
	if n, _ := db.Category.Count("id"); n > 0 {
		return
	}
	cats := []*models.Category{
		{Name: "综合讨论", Slug: "general", Description: "欢迎畅所欲言", Color: "#718096", Position: 0},
		{Name: "技术交流", Slug: "tech", Description: "编程与技术话题", Color: "#3182ce", Position: 1},
	}
	for _, cat := range cats {
		if err := db.Category.Insert().One(cat); err != nil {
			log.Printf("Failed to seed category %s: %v", cat.Name, err)
		}
	}
}

// seedDemo 只在没有任何话题时写入演示数据：确保每个话题都带标签，
// 让热门标签侧栏与“每个帖子要带 tag”的需求有可测试的数据。
func seedDemo() {
	db := models.GetDB()
	if n, _ := db.Topic.Count("id"); n > 0 {
		return
	}

	// 演示用户：alice（作者）、bob（回复）
	const demoPass = "password"
	getUser := func(username, email string) *models.User {
		if u, _ := db.User.Select().Where("username = ?", username).One(); u != nil {
			return u
		}
		u := &models.User{Username: username, Email: email, Role: models.RoleUser, IsActive: true}
		if err := u.SetPassword(demoPass); err == nil {
			if err := db.User.Insert().One(u); err != nil {
				log.Printf("Failed to seed user %s: %v", username, err)
				return nil
			}
		}
		return u
	}
	alice := getUser("alice", "alice@flock.local")
	authorID := int64(1)
	if alice != nil {
		authorID = alice.ID
	}
	bob := getUser("bob", "bob@flock.local")
	readerID := authorID
	if bob != nil {
		readerID = bob.ID
	}

	cats := map[string]*models.Category{}
	if list, err := db.Category.Select().All(); err == nil {
		for _, c := range list {
			cats[c.Slug] = c
		}
	}

	demos := []struct {
		title string
		cat   string
		tags  []string
		body  string
		reply string
	}{
		{
			title: "用 Go 写一个极简论坛的心得",
			cat:   "tech",
			tags:  []string{"Go", "后端", "开发"},
			body:  "最近用 Go 的 Echo + goent 写了个极简论坛，分享几点体会。\n\n## 目录\n\n先交代一下整体结构：\n\n- 前后端分离，前端是 Alpine.js + 原生 CSS\n- 后端用 Echo 提供 RESTful API\n- 数据层用 goent 做了轻量 ORM\n\n## 遇到的坑\n\n1. 单回车换行需要特殊的解析\n2. 标签的热门排序要基于话题计数\n\n这些问题都在社区里找到了答案，值得记录下来。\n\n> 感谢大家的讨论……",
			reply: "学习了，请问 goent 相比 GORM 上手成本如何？",
		},
		{
			title: "水粉画风格的配色怎么选才简洁",
			cat:   "general",
			tags:  []string{"设计", "配色"},
			body:  "想给论坛加点水粉画（flock）的柔和质感，但又不能太花。\n\n## 思路\n\n水粉画的精髓是“柔和而沉稳”，颜色饱和度低、有纸张的质感。\n\n第一步先把背景做成淡淡的纸色，再在角落加一点晕染。\n\n## 总结\n\n少即是多，留白和质感比鲜艳更重要。",
			reply: "",
		},
		{
			title: "前端状态管理与 URL 参数同步",
			cat:   "tech",
			tags:  []string{"前端", "状态管理", "路由"},
			body:  "让分类和标签的选中状态跟随 URL，刷新后不丢失。\n\n## 做法\n\n用 useSearchParams 读取 cat/tag，选中项回写回 URL。\n\n这样无论是首页还是话题详情，左侧栏都保持一致。",
			reply: "这个方法很实用，学到了。",
		},
		{
			title: "Markdown 里的换行到底怎么处理",
			cat:   "tech",
			tags:  []string{"Markdown", "规范"},
			body:  "很多地方单回车不会换行，需要额外处理。\n\n第一行\n第二行\n第三行\n\n## 结论\n\n配合 remark-breaks 后，单回车就能转成 `<br>` 了。",
			reply: "",
		},
		{
			title: "如何给长帖子加一个右侧目录",
			cat:   "general",
			tags:  []string{"UI", "体验"},
			body:  "帖子太长想加目录，方便读者跳转。\n\n## 实现\n\n提取 Markdown 里的标题生成锚点。\n\n右侧目录在内容很多时自动出现滚动条。\n\n## 效果\n\n阅读长文体验好了不少。",
			reply: "",
		},
	}

	for _, d := range demos {
		topic := &models.Topic{
			Title:      d.title,
			Slug:       helpers.Slugify(d.title),
			CategoryID: nil,
			UserID:     authorID,
		}
		if c, ok := cats[d.cat]; ok {
			id := c.ID
			topic.CategoryID = &id
		}
		now := time.Now()
		err := db.BeginTransaction(func(tx model.Transaction) error {
			if err := db.Topic.Insert().OnTransaction(tx).One(topic); err != nil {
				return err
			}
			first := &models.Post{TopicID: topic.ID, UserID: authorID, PostNumber: 1, Raw: d.body}
			if err := db.Post.Insert().OnTransaction(tx).One(first); err != nil {
				return err
			}
			lastID := first.ID
			posts := int64(1)
			if d.reply != "" && readerID != authorID {
				rep := &models.Post{TopicID: topic.ID, UserID: readerID, PostNumber: 2, Raw: d.reply}
				if err := db.Post.Insert().OnTransaction(tx).One(rep); err != nil {
					return err
				}
				lastID = rep.ID
				posts = 2
			}
			if err := db.Topic.Update().
				Set(
					goent.Pair{Key: "posts_count", Value: posts},
					goent.Pair{Key: "last_post_id", Value: lastID},
					goent.Pair{Key: "last_posted_at", Value: now},
					goent.Pair{Key: "updated_at", Value: now},
				).
				Filter(goent.Equals(db.Topic.Field("id"), topic.ID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
			for _, name := range d.tags {
				tag, err := seedTag(tx, name)
				if err != nil {
					return err
				}
				tt := &models.TopicTag{TopicID: topic.ID, TagID: tag.ID}
				if err := db.TopicTag.Insert().OnTransaction(tx).One(tt); err != nil {
					return err
				}
				if err := db.Tag.Update().
					Set(goent.Pair{Key: "topics_count", Value: tag.TopicsCount + 1}).
					Filter(goent.Equals(db.Tag.Field("id"), tag.ID)).
					OnTransaction(tx).
					Exec(); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("Failed to seed demo topic %q: %v", d.title, err)
		}
	}
	log.Printf("Seeded %d demo topics with tags", len(demos))
}

// seedTag 复用（或新建）一个标签并返回，供演示数据使用。
func seedTag(tx model.Transaction, name string) (*models.Tag, error) {
	db := models.GetDB()
	name = strings.ToLower(strings.TrimSpace(name))
	slug := helpers.Slugify(name)
	if tag, err := db.Tag.Select().
		Where("slug = ? AND name = ?", slug, name).
		OnTransaction(tx).One(); err == nil && tag != nil {
		return tag, nil
	}
	tag := &models.Tag{Name: name, Slug: slug}
	return tag, db.Tag.Insert().OnTransaction(tx).One(tag)
}
