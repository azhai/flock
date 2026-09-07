package handlers

import (
	"strconv"

	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
	"github.com/azhai/gouache/helpers"
	"github.com/azhai/gouache/middleware"
	"github.com/azhai/gouache/models"
	"github.com/labstack/echo/v5"
)

// ToggleBookmark 收藏/取消收藏话题，返回当前状态
func ToggleBookmark(c *echo.Context) error {
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

	existing, _ := db.Bookmark.Select().
		Where("user_id = ? AND topic_id = ?", user.ID, topic.ID).
		One()

	var bookmarked bool
	if existing != nil {
		err = db.BeginTransaction(func(tx model.Transaction) error {
			return db.Bookmark.Delete().
				Filter(goent.Equals(db.Bookmark.Field("id"), existing.ID)).
				OnTransaction(tx).
				Exec()
		})
	} else {
		bookmarked = true
		err = db.BeginTransaction(func(tx model.Transaction) error {
			bm := &models.Bookmark{UserID: user.ID, TopicID: topic.ID}
			return db.Bookmark.Insert().OnTransaction(tx).One(bm)
		})
	}
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	return c.JSON(200, map[string]any{"bookmarked": bookmarked})
}

// ListBookmarks 当前用户的收藏话题列表
func ListBookmarks(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	db := models.GetDB()

	bms, err := db.Bookmark.Select().
		Where("user_id = ?", user.ID).
		OrderBy("id DESC").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载收藏失败"))
	}
	ids := make([]int64, 0, len(bms))
	for _, bm := range bms {
		ids = append(ids, bm.TopicID)
	}
	if len(ids) == 0 {
		return c.JSON(200, []*TopicResponse{})
	}

	topics, err := db.Topic.Select().
		Filter(goent.In(db.Topic.Field("id"), ids)).
		With("User", "Category").
		OrderBy("last_posted_at DESC", "id DESC").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载收藏失败"))
	}

	tagSet := loadTopicTags(ids)
	resp := make([]*TopicResponse, 0, len(topics))
	for _, t := range topics {
		t.Tags = tagSet[t.ID]
		t.Bookmarked = true
		resp = append(resp, ToTopicResponse(t, 0))
	}
	return c.JSON(200, resp)
}
