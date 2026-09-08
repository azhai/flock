package handlers

import (
	"strings"

	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/models"
	"github.com/azhai/goent"
	"github.com/labstack/echo/v5"
)

type CategoryRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Position    *int   `json:"position"`
	ParentID    *int64 `json:"parent_id"`
}

// ListCategories 返回嵌套分类树（含各分类直属话题数）
func ListCategories(c *echo.Context) error {
	db := models.GetDB()
	cats, err := db.Category.Select().OrderBy("position", "id").All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载分类失败"))
	}
	counts := make(map[int64]int64, len(cats))
	for _, cat := range cats {
		n, _ := db.Topic.Filter(goent.Equals(db.Topic.Field("category_id"), cat.ID)).Count("id")
		counts[cat.ID] = n
	}
	return c.JSON(200, buildCategoryTree(cats, counts))
}

func buildCategoryTree(cats []*models.Category, counts map[int64]int64) []*CategoryResponse {
	nodes := make(map[int64]*CategoryResponse, len(cats))
	for _, cat := range cats {
		nodes[cat.ID] = ToCategoryResponse(cat, counts[cat.ID])
	}
	roots := make([]*CategoryResponse, 0, len(cats))
	for _, cat := range cats {
		node := nodes[cat.ID]
		if cat.ParentID != nil {
			if parent, ok := nodes[*cat.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots
}

// CreateCategory 新建分类（admin）
func CreateCategory(c *echo.Context) error {
	var req CategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.JSON(400, helpers.ErrorJSON("分类名称不能为空"))
	}
	slug := req.Slug
	if slug == "" {
		slug = helpers.Slugify(req.Name)
	}

	db := models.GetDB()
	if cat, _ := db.Category.Select().Where("slug = ?", slug).One(); cat != nil {
		return c.JSON(409, helpers.ErrorJSON("slug 已存在"))
	}

	cat := &models.Category{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Color:       req.Color,
		ParentID:    req.ParentID,
	}
	if req.Position != nil {
		cat.Position = *req.Position
	}
	if err := db.Category.Insert().One(cat); err != nil {
		return c.JSON(500, helpers.ErrorJSON("创建分类失败"))
	}
	return c.JSON(201, ToCategoryResponse(cat, 0))
}

// UpdateCategory 更新分类（admin）
func UpdateCategory(c *echo.Context) error {
	cat, err := findCategory(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("分类不存在"))
	}
	var req CategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	req.Name = strings.TrimSpace(req.Name)

	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Slug != "" && req.Slug != cat.Slug {
		db := models.GetDB()
		if existed, _ := db.Category.Select().Where("slug = ?", req.Slug).One(); existed != nil {
			return c.JSON(409, helpers.ErrorJSON("slug 已存在"))
		}
		cat.Slug = req.Slug
	}
	cat.Description = req.Description
	if req.Color != "" {
		cat.Color = req.Color
	}
	if req.Position != nil {
		cat.Position = *req.Position
	}
	if req.ParentID != nil {
		if *req.ParentID == cat.ID {
			return c.JSON(400, helpers.ErrorJSON("分类不能作为自身的父分类"))
		}
		cat.ParentID = req.ParentID
	}
	cat.UpdatedAt = timeNow()

	if err := models.GetDB().Category.Save().One(cat); err != nil {
		return c.JSON(500, helpers.ErrorJSON("更新分类失败"))
	}
	return c.JSON(200, ToCategoryResponse(cat, 0))
}

// DeleteCategory 删除分类（admin），仅当无子分类且无直属话题
func DeleteCategory(c *echo.Context) error {
	cat, err := findCategory(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("分类不存在"))
	}
	db := models.GetDB()
	if child, _ := db.Category.Select().Where("parent_id = ?", cat.ID).One(); child != nil {
		return c.JSON(400, helpers.ErrorJSON("请先删除子分类"))
	}
	if n, _ := db.Topic.Filter(goent.Equals(db.Topic.Field("category_id"), cat.ID)).Count("id"); n > 0 {
		return c.JSON(400, helpers.ErrorJSON("该分类下仍有话题，无法删除"))
	}
	if err := db.Category.Delete().Filter(goent.Equals(db.Category.Field("id"), cat.ID)).Exec(); err != nil {
		return c.JSON(500, helpers.ErrorJSON("删除分类失败"))
	}
	return c.JSON(200, map[string]any{"message": "deleted"})
}

func findCategory(c *echo.Context) (*models.Category, error) {
	idStr := c.Param("id")
	if idStr != "" {
		cat, err := models.GetDB().Category.Select().Where("id = ?", idStr).One()
		if err == nil {
			return cat, nil
		}
	}
	slug := c.Param("slug")
	return models.GetDB().Category.Select().Where("slug = ?", slug).One()
}
