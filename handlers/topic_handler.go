package handlers

import (
	"strconv"
	"strings"

	"github.com/azhai/flock/config"
	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/middleware"
	"github.com/azhai/flock/models"
	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
	"github.com/labstack/echo/v5"
)

type CreateTopicRequest struct {
	Title      string   `json:"title"`
	Raw        string   `json:"raw"`
	CategoryID *int64   `json:"category_id"`
	Tags       []string `json:"tags"`
}

type UpdateTopicRequest struct {
	Title      string   `json:"title"`
	CategoryID *int64   `json:"category_id"`
	Tags       []string `json:"tags"`
}

// ListTopics 话题列表：sort=latest|top|likes|hot，支持 category_id 与 q 过滤
func ListTopics(c *echo.Context) error {
	db := models.GetDB()
	p := helpers.GetPagination(c)

	query := db.Topic.Select()

	// 全文检索分支：FTS5(trigram) 覆盖标题+正文，按相关度排序
	if fq := strings.TrimSpace(c.QueryParam("q")); fq != "" {
		total, ids := models.SearchTopics(db, fq, helpers.GetOffset(p.Page, p.PerPage), p.PerPage)
		topics, _ := db.Topic.Select().
			Filter(goent.In(db.Topic.Field("id"), ids)).
			With("User", "Category").
			All()
		// 保持 FTS 相关度顺序
		byID := make(map[int64]*models.Topic, len(topics))
		for _, t := range topics {
			byID[t.ID] = t
		}
		ordered := make([]*models.Topic, 0, len(ids))
		for _, id := range ids {
			if t, ok := byID[id]; ok {
				ordered = append(ordered, t)
			}
		}
		return c.JSON(200, buildTopicsResponse(c, db, ordered, total, p))
	}

	if cid := strings.TrimSpace(c.QueryParam("category_id")); cid != "" {
		if id, err := strconv.ParseInt(cid, 10, 64); err == nil {
			query = query.Where("category_id = ?", id)
		}
	}
	if q := strings.TrimSpace(c.QueryParam("q")); q != "" {
		query = query.Where("title LIKE ?", "%"+q+"%")
	}
	if tagParam := strings.TrimSpace(c.QueryParam("tag")); tagParam != "" {
		tag, err := db.Tag.Select().
			Filter(goent.Or(
				goent.Equals(db.Tag.Field("slug"), tagParam),
				goent.Equals(db.Tag.Field("name"), tagParam),
			)).One()
		if err != nil || tag == nil {
			return c.JSON(200, helpers.NewPaginatedResponse([]*TopicResponse{}, p.Page, p.PerPage, 0))
		}
		tts, _ := db.TopicTag.Select().
			Filter(goent.Equals(db.TopicTag.Field("tag_id"), tag.ID)).
			All()
		ids := make([]int64, 0, len(tts))
		for _, tt := range tts {
			ids = append(ids, tt.TopicID)
		}
		if len(ids) == 0 {
			return c.JSON(200, helpers.NewPaginatedResponse([]*TopicResponse{}, p.Page, p.PerPage, 0))
		}
		query = query.Filter(goent.In(db.Topic.Field("id"), ids))
	}

	sortCol := map[string]string{
		config.SortTop:   "reply_count",
		config.SortLikes: "like_count",
		config.SortHot:   "last_posted_at",
	}[c.QueryParam("sort")]
	if sortCol == "" {
		sortCol = "created_at"
	}

	total, _ := query.Count("id")
	topics, err := query.
		OrderBy("pinned DESC", sortCol+" DESC", "id DESC").
		Skip(helpers.GetOffset(p.Page, p.PerPage)).
		Take(p.PerPage).
		With("User", "Category").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载话题失败"))
	}
	return c.JSON(200, buildTopicsResponse(c, db, topics, total, p))
}

// buildTopicsResponse 组装话题列表响应：标签/参与者/收藏/关注 + 分页信息。
func buildTopicsResponse(c *echo.Context, db *models.Database, topics []*models.Topic, total int64, p helpers.PaginationParams) *helpers.Response {
	tagSet := make(map[int64][]*models.Tag)
	parts := make(map[int64][]*models.User)
	if len(topics) > 0 {
		ids := make([]int64, len(topics))
		for i, t := range topics {
			ids[i] = t.ID
		}
		tagSet = loadTopicTags(ids)
		parts = loadParticipants(ids)
	}

	bookmarks := map[int64]bool{}
	followed := map[int64]bool{}
	if userID := middleware.GetCurrentUserID(c); userID != 0 {
		ids := make([]int64, len(topics))
		for i, t := range topics {
			ids[i] = t.ID
		}
		bookmarks = loadBookmarks(ids, userID)
		followed = loadFollowed(ids, userID)
	}

	resp := make([]*TopicResponse, 0, len(topics))
	for _, t := range topics {
		t.Tags = tagSet[t.ID]
		t.Participants = parts[t.ID]
		t.Bookmarked = bookmarks[t.ID]
		t.Followed = followed[t.ID]
		resp = append(resp, ToTopicResponse(t, 0))
	}
	return helpers.NewPaginatedResponse(resp, p.Page, p.PerPage, total)
}

// GetTopic 话题详情，浏览量 +1
func GetTopic(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid topic id"))
	}
	db := models.GetDB()
	topic, err := db.Topic.Select().Where("id = ?", id).With("User", "Category").One()
	if err != nil || topic == nil {
		return c.JSON(404, helpers.ErrorJSON("话题不存在"))
	}
	_ = db.Topic.Update().
		Set(goent.Pair{Key: "views_count", Value: topic.ViewsCount + 1}).
		Filter(goent.Equals(db.Topic.Field("id"), topic.ID)).
		Exec()
	topic.ViewsCount++

	topic.Tags = loadTopicTags([]int64{topic.ID})[topic.ID]
	topic.Participants = loadParticipants([]int64{topic.ID})[topic.ID]
	if userID := middleware.GetCurrentUserID(c); userID != 0 {
		topic.Bookmarked = loadBookmarks([]int64{topic.ID}, userID)[topic.ID]
		topic.Followed = loadFollowed([]int64{topic.ID}, userID)[topic.ID]
	}

	return c.JSON(200, ToTopicResponse(topic, 0))
}

// CreateTopic 发新话题：事务创建话题 + 首帖
func CreateTopic(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	var req CreateTopicRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return c.JSON(400, helpers.ErrorJSON("标题不能为空"))
	}
	if strings.TrimSpace(req.Raw) == "" {
		return c.JSON(400, helpers.ErrorJSON("内容不能为空"))
	}

	db := models.GetDB()
	topic := &models.Topic{
		Title:      req.Title,
		Slug:       helpers.Slugify(req.Title),
		CategoryID: req.CategoryID,
		UserID:     user.ID,
	}

	err := db.BeginTransaction(func(tx model.Transaction) error {
		if err := db.Topic.Insert().OnTransaction(tx).One(topic); err != nil {
			return err
		}
		post := &models.Post{
			TopicID:    topic.ID,
			UserID:     user.ID,
			PostNumber: 1,
			Raw:        req.Raw,
		}
		if err := db.Post.Insert().OnTransaction(tx).One(post); err != nil {
			return err
		}
		then := timeNow()
		if err := db.Topic.Update().
			Set(
				goent.Pair{Key: "posts_count", Value: 1},
				goent.Pair{Key: "last_post_id", Value: post.ID},
				goent.Pair{Key: "last_posted_at", Value: then},
				goent.Pair{Key: "updated_at", Value: then},
			).
			Filter(goent.Equals(db.Topic.Field("id"), topic.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		// 关联标签（复用/新建），并累加各标签 topic 计数
		topic.Tags = nil
		seen := map[string]bool{}
		for _, name := range req.Tags {
			name = strings.TrimSpace(name)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			tag, err := upsertTag(tx, name)
			if err != nil {
				return err
			}
			tt := &models.TopicTag{TopicID: topic.ID, TagID: tag.ID}
			if err := db.TopicTag.Insert().OnTransaction(tx).One(tt); err != nil {
				return err
			}
			if err := db.Tag.Update().
				Set(goent.Pair{Key: "topics_count", Value: maxInt(tag.TopicsCount+1, 0)}).
				Filter(goent.Equals(db.Tag.Field("id"), tag.ID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
			tag.TopicsCount++
			topic.Tags = append(topic.Tags, tag)
		}
		return nil
	})
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("创建话题失败"))
	}

	created, _ := db.Topic.Select().Where("id = ?", topic.ID).With("User", "Category").One()
	created.Tags = topic.Tags
	created.Followed = true // 作者默认关注自己的话题
	// 首帖里的 @提及 也触发通知
	if first, err := db.Post.Select().
		Where("topic_id = ? AND post_number = ?", topic.ID, 1).
		One(); err == nil && first != nil {
		createMentionNotifications(first, user.ID)
	}
	return c.JSON(201, ToTopicResponse(created, 0))
}

// ToggleFollow 关注/取消关注话题，返回当前状态
func ToggleFollow(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid topic id"))
	}
	db := models.GetDB()
	topic, err := db.Topic.Select().Where("id = ?", topicID).One()
	if err != nil || topic == nil {
		return c.JSON(404, helpers.ErrorJSON("话题不存在"))
	}

	existing, _ := db.TopicFollower.Select().
		Where("user_id = ? AND topic_id = ?", user.ID, topic.ID).
		One()

	var followed bool
	if existing != nil {
		err = db.BeginTransaction(func(tx model.Transaction) error {
			return db.TopicFollower.Delete().
				Filter(goent.Equals(db.TopicFollower.Field("id"), existing.ID)).
				OnTransaction(tx).
				Exec()
		})
	} else {
		followed = true
		err = db.BeginTransaction(func(tx model.Transaction) error {
			f := &models.TopicFollower{UserID: user.ID, TopicID: topic.ID}
			return db.TopicFollower.Insert().OnTransaction(tx).One(f)
		})
	}
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	return c.JSON(200, map[string]any{"followed": followed})
}

// UpdateTopic 更新话题（作者或 staff）
func UpdateTopic(c *echo.Context) error {
	topic, err := findTopic(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("话题不存在"))
	}
	if !canEditTopic(c, topic) {
		return c.JSON(403, helpers.ErrorJSON("无权限操作该话题"))
	}
	var req UpdateTopicRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	req.Title = strings.TrimSpace(req.Title)

	db := models.GetDB()
	err = db.BeginTransaction(func(tx model.Transaction) error {
		if req.Title != "" {
			topic.Title = req.Title
			topic.Slug = helpers.Slugify(req.Title)
		}
		if req.CategoryID != nil {
			topic.CategoryID = req.CategoryID
		}
		topic.UpdatedAt = timeNow()
		if err := db.Topic.Save().OnTransaction(tx).One(topic); err != nil {
			return err
		}
		if req.Tags != nil {
			if err := replaceTopicTags(tx, topic.ID, req.Tags); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("更新话题失败"))
	}
	topic.Tags = loadTopicTags([]int64{topic.ID})[topic.ID]
	return c.JSON(200, ToTopicResponse(topic, 0))
}

// replaceTopicTags 以给定标签集重建某话题的标签关联，并校正各 tag 计数（事务内调用）。
func replaceTopicTags(tx model.Transaction, topicID int64, names []string) error {
	db := models.GetDB()
	// 1) 清空旧关联并回减计数
	tts, err := db.TopicTag.Select().Where("topic_id = ?", topicID).OnTransaction(tx).All()
	if err != nil {
		return err
	}
	for _, tt := range tts {
		tag, err := db.Tag.Select().Where("id = ?", tt.TagID).OnTransaction(tx).One()
		if err != nil {
			return err
		}
		if err := db.Tag.Update().
			Set(goent.Pair{Key: "topics_count", Value: maxInt(tag.TopicsCount-1, 0)}).
			Filter(goent.Equals(db.Tag.Field("id"), tt.TagID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		if err := db.TopicTag.Delete().
			Filter(goent.Equals(db.TopicTag.Field("id"), tt.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
	}
	// 2) 重建关联（复用/新建标签）
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		tag, err := upsertTag(tx, name)
		if err != nil {
			return err
		}
		tt := &models.TopicTag{TopicID: topicID, TagID: tag.ID}
		if err := db.TopicTag.Insert().OnTransaction(tx).One(tt); err != nil {
			return err
		}
		if err := db.Tag.Update().
			Set(goent.Pair{Key: "topics_count", Value: maxInt(tag.TopicsCount+1, 0)}).
			Filter(goent.Equals(db.Tag.Field("id"), tag.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTopic 删除话题并级联删除帖子与点赞（作者或 staff）
func DeleteTopic(c *echo.Context) error {
	topic, err := findTopic(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("话题不存在"))
	}
	if !canEditTopic(c, topic) {
		return c.JSON(403, helpers.ErrorJSON("无权限操作该话题"))
	}

	db := models.GetDB()
	err = db.BeginTransaction(func(tx model.Transaction) error {
		// 清理标签关联并回减计数
		tts, err := db.TopicTag.Select().Where("topic_id = ?", topic.ID).OnTransaction(tx).All()
		if err != nil {
			return err
		}
		for _, tt := range tts {
			tag, err := db.Tag.Select().Where("id = ?", tt.TagID).OnTransaction(tx).One()
			if err != nil {
				return err
			}
			if err := db.Tag.Update().
				Set(goent.Pair{Key: "topics_count", Value: maxInt(tag.TopicsCount-1, 0)}).
				Filter(goent.Equals(db.Tag.Field("id"), tt.TagID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
		}
		if err := db.TopicTag.Delete().
			Filter(goent.Equals(db.TopicTag.Field("topic_id"), topic.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		if err := db.Bookmark.Delete().
			Filter(goent.Equals(db.Bookmark.Field("topic_id"), topic.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		posts, err := db.Post.Select().Where("topic_id = ?", topic.ID).OnTransaction(tx).All()
		if err != nil {
			return err
		}
		for _, p := range posts {
			if err := db.PostLike.Delete().
				Filter(goent.Equals(db.PostLike.Field("post_id"), p.ID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
			if err := db.PostVote.Delete().
				Filter(goent.Equals(db.PostVote.Field("post_id"), p.ID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
			if err := db.Report.Delete().
				Filter(goent.Equals(db.Report.Field("post_id"), p.ID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
		}
		if err := db.TopicFollower.Delete().
			Filter(goent.Equals(db.TopicFollower.Field("topic_id"), topic.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		if err := db.Notification.Delete().
			Filter(goent.Or(
				goent.Equals(db.Notification.Field("topic_id"), topic.ID),
				goent.In(db.Notification.Field("post_id"), postIDs(posts)),
			)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		if err := db.Post.Delete().
			Filter(goent.Equals(db.Post.Field("topic_id"), topic.ID)).
			OnTransaction(tx).
			Exec(); err != nil {
			return err
		}
		return db.Topic.Delete().
			Filter(goent.Equals(db.Topic.Field("id"), topic.ID)).
			OnTransaction(tx).
			Exec()
	})
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("删除话题失败"))
	}
	return c.JSON(200, map[string]any{"message": "deleted"})
}

// PinTopic / UnpinTopic 置顶（staff）
func PinTopic(c *echo.Context) error { return setTopicFlag(c, "pinned", true) }
func UnpinTopic(c *echo.Context) error {
	return setTopicFlag(c, "pinned", false)
}

// CloseTopic / OpenTopic 锁定（staff）
func CloseTopic(c *echo.Context) error { return setTopicFlag(c, "closed", true) }
func OpenTopic(c *echo.Context) error  { return setTopicFlag(c, "closed", false) }

func setTopicFlag(c *echo.Context, field string, val bool) error {
	topic, err := findTopic(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("话题不存在"))
	}
	if !middleware.IsStaff(c) {
		return c.JSON(403, helpers.ErrorJSON("需要管理员或版主权限"))
	}
	db := models.GetDB()
	if err := db.Topic.Update().
		Set(goent.Pair{Key: field, Value: val}, goent.Pair{Key: "updated_at", Value: timeNow()}).
		Filter(goent.Equals(db.Topic.Field("id"), topic.ID)).
		Exec(); err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	return c.JSON(200, map[string]any{"message": "ok"})
}

func findTopic(c *echo.Context) (*models.Topic, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}
	return models.GetDB().Topic.Select().Where("id = ?", id).One()
}

func canEditTopic(c *echo.Context, topic *models.Topic) bool {
	if middleware.IsStaff(c) {
		return true
	}
	return middleware.GetCurrentUserID(c) == topic.UserID
}

// loadFollowed 批量查询某用户关注了哪些话题，返回 topic_id → bool。
func loadFollowed(topicIDs []int64, userID int64) map[int64]bool {
	out := make(map[int64]bool, len(topicIDs))
	if len(topicIDs) == 0 || userID == 0 {
		return out
	}
	db := models.GetDB()
	fs, _ := db.TopicFollower.Select().
		Filter(goent.And(
			goent.Equals(db.TopicFollower.Field("user_id"), userID),
			goent.In(db.TopicFollower.Field("topic_id"), topicIDs),
		)).
		All()
	for _, f := range fs {
		out[f.TopicID] = true
	}
	return out
}

// upsertTag 按 slug 复用或新建标签，返回标签（事务内调用）。
func upsertTag(tx model.Transaction, name string) (*models.Tag, error) {
	db := models.GetDB()
	name = strings.ToLower(strings.TrimSpace(name))
	slug := helpers.Slugify(name)

	tag, err := db.Tag.Select().Where("slug = ? AND name = ?", slug, name).OnTransaction(tx).One()
	if err == nil && tag != nil {
		return tag, nil
	}
	tag = &models.Tag{Name: name, Slug: slug}
	if err := db.Tag.Insert().OnTransaction(tx).One(tag); err != nil {
		// 并发/已存在则重查
		existing, e2 := db.Tag.Select().Where("slug = ?", slug).OnTransaction(tx).One()
		if e2 == nil && existing != nil {
			return existing, nil
		}
		return nil, err
	}
	return tag, nil
}

// loadTopicTags 批量加载若干话题的标签，返回 topic_id → []Tag。
func loadTopicTags(topicIDs []int64) map[int64][]*models.Tag {
	db := models.GetDB()
	out := make(map[int64][]*models.Tag, len(topicIDs))
	if len(topicIDs) == 0 {
		return out
	}
	tts, _ := db.TopicTag.Select().
		Filter(goent.In(db.TopicTag.Field("topic_id"), topicIDs)).
		All()
	tagIDs := make([]int64, 0, len(tts))
	seen := map[int64]bool{}
	for _, tt := range tts {
		if !seen[tt.TagID] {
			seen[tt.TagID] = true
			tagIDs = append(tagIDs, tt.TagID)
		}
	}
	tags, _ := db.Tag.Select().
		Filter(goent.In(db.Tag.Field("id"), tagIDs)).
		All()
	byID := make(map[int64]*models.Tag, len(tags))
	for _, t := range tags {
		byID[t.ID] = t
	}
	for _, tt := range tts {
		if tg, ok := byID[tt.TagID]; ok {
			out[tt.TopicID] = append(out[tt.TopicID], tg)
		}
	}
	return out
}

// loadParticipants 批量查询话题参与者（去重用户），按首帖作者在前返回 topic_id → []User。
func loadParticipants(topicIDs []int64) map[int64][]*models.User {
	out := make(map[int64][]*models.User, len(topicIDs))
	if len(topicIDs) == 0 {
		return out
	}
	db := models.GetDB()
	posts, _ := db.Post.Select().
		Filter(goent.In(db.Post.Field("topic_id"), topicIDs)).
		OrderBy("post_number ASC").
		All()

	userSeen := make(map[int64]bool)
	topicUsers := make(map[int64][]int64)
	userIDs := make([]int64, 0)
	for _, p := range posts {
		if !userSeen[p.UserID] {
			userSeen[p.UserID] = true
			userIDs = append(userIDs, p.UserID)
		}
		topicUsers[p.TopicID] = append(topicUsers[p.TopicID], p.UserID)
	}
	if len(userIDs) == 0 {
		return out
	}
	users, _ := db.User.Select().
		Filter(goent.In(db.User.Field("id"), userIDs)).
		All()
	byID := make(map[int64]*models.User, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}
	for tid, ids := range topicUsers {
		seen := make(map[int64]bool)
		for _, uid := range ids {
			if seen[uid] {
				continue
			}
			seen[uid] = true
			if u, ok := byID[uid]; ok {
				out[tid] = append(out[tid], u)
			}
		}
	}
	return out
}

// loadBookmarks 批量查询某用户收藏了哪些话题，返回 topic_id → bool。
func loadBookmarks(topicIDs []int64, userID int64) map[int64]bool {
	out := make(map[int64]bool, len(topicIDs))
	if len(topicIDs) == 0 || userID == 0 {
		return out
	}
	db := models.GetDB()
	bms, _ := db.Bookmark.Select().
		Filter(goent.And(
			goent.Equals(db.Bookmark.Field("user_id"), userID),
			goent.In(db.Bookmark.Field("topic_id"), topicIDs),
		)).
		All()
	for _, bm := range bms {
		out[bm.TopicID] = true
	}
	return out
}
