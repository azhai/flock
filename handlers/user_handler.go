package handlers

import (
	"github.com/azhai/goent"
	"github.com/azhai/gouache/helpers"
	"github.com/azhai/gouache/middleware"
	"github.com/azhai/gouache/models"
	"github.com/labstack/echo/v5"
)

// GetUser 用户主页（公开信息 + 话题/帖子统计）
func GetUser(c *echo.Context) error {
	username := c.Param("username")
	db := models.GetDB()
	user, _ := db.User.Select().Where("username = ?", username).One()
	if user == nil {
		return c.JSON(404, helpers.ErrorJSON("用户不存在"))
	}
	topics, _ := db.Topic.Filter(goent.Equals(db.Topic.Field("user_id"), user.ID)).Count("id")
	posts, _ := db.Post.Filter(goent.Equals(db.Post.Field("user_id"), user.ID)).Count("id")
	resp := ToUserResponse(user)
	resp.TopicsCount = topics
	resp.PostsCount = posts
	return c.JSON(200, resp)
}

// UpdateProfile 更新自己资料
func UpdateProfile(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	var req struct {
		FullName string `json:"full_name"`
		Bio      string `json:"bio"`
		Website  string `json:"website"`
		Location string `json:"location"`
		Avatar   string `json:"avatar"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	db := models.GetDB()
	user.FullName = req.FullName
	user.Bio = req.Bio
	user.Website = req.Website
	user.Location = req.Location
	user.Avatar = req.Avatar
	user.UpdatedAt = timeNow()
	if err := db.User.Save().One(user); err != nil {
		return c.JSON(500, helpers.ErrorJSON("更新资料失败"))
	}
	return c.JSON(200, ToUserResponse(user))
}

// GetUserTopics 某用户发表的话题
func GetUserTopics(c *echo.Context) error {
	username := c.Param("username")
	db := models.GetDB()
	user, _ := db.User.Select().Where("username = ?", username).One()
	if user == nil {
		return c.JSON(404, helpers.ErrorJSON("用户不存在"))
	}
	p := helpers.GetPagination(c)
	total, _ := db.Topic.Filter(goent.Equals(db.Topic.Field("user_id"), user.ID)).Count("id")
	topics, err := db.Topic.Select().
		Where("user_id = ?", user.ID).
		OrderBy("id DESC").
		Skip(helpers.GetOffset(p.Page, p.PerPage)).
		Take(p.PerPage).
		With("Category").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载话题失败"))
	}
	resp := make([]*TopicResponse, 0, len(topics))
	for _, t := range topics {
		resp = append(resp, ToTopicResponse(t, 0))
	}
	return c.JSON(200, helpers.NewPaginatedResponse(resp, p.Page, p.PerPage, total))
}

// GetUserPosts 某用户发表的帖子
func GetUserPosts(c *echo.Context) error {
	username := c.Param("username")
	db := models.GetDB()
	user, _ := db.User.Select().Where("username = ?", username).One()
	if user == nil {
		return c.JSON(404, helpers.ErrorJSON("用户不存在"))
	}
	p := helpers.GetPagination(c)
	total, _ := db.Post.Filter(goent.Equals(db.Post.Field("user_id"), user.ID)).Count("id")
	posts, err := db.Post.Select().
		Where("user_id = ?", user.ID).
		OrderBy("id DESC").
		Skip(helpers.GetOffset(p.Page, p.PerPage)).
		Take(p.PerPage).
		With("User").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载帖子失败"))
	}
	resp := make([]*PostResponse, 0, len(posts))
	for _, post := range posts {
		resp = append(resp, ToPostResponse(post, false))
	}
	return c.JSON(200, helpers.NewPaginatedResponse(resp, p.Page, p.PerPage, total))
}
