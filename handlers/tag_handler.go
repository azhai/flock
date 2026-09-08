package handlers

import (
	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/models"
	"github.com/labstack/echo/v5"
)

// ListTags 标签列表，按话题计数降序（只读，公开）
func ListTags(c *echo.Context) error {
	db := models.GetDB()
	tags, err := db.Tag.Select().
		Where("topics_count > 0").
		OrderBy("topics_count DESC", "id DESC").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载标签失败"))
	}
	resp := make([]*TagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, ToTagResponse(t))
	}
	return c.JSON(200, resp)
}
