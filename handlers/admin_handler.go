package handlers

import (
	"github.com/azhai/goent"
	"github.com/azhai/gouache/helpers"
	"github.com/azhai/gouache/models"
	"github.com/labstack/echo/v5"
)

// ListUsers 用户列表（admin）
func ListUsers(c *echo.Context) error {
	db := models.GetDB()
	p := helpers.GetPagination(c)
	q := c.QueryParam("q")

	query := db.User.Select()
	if q != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	total, _ := query.Count("id")
	users, err := query.
		OrderBy("id ASC").
		Skip(helpers.GetOffset(p.Page, p.PerPage)).
		Take(p.PerPage).
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载用户失败"))
	}

	resp := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, ToUserResponse(u))
	}
	return c.JSON(200, helpers.NewPaginatedResponse(resp, p.Page, p.PerPage, total))
}

// AdminUpdateUser 管理员更新用户：角色、状态、密码、昵称
func AdminUpdateUser(c *echo.Context) error {
	username := c.Param("username")
	db := models.GetDB()
	user, _ := db.User.Select().Where("username = ?", username).One()
	if user == nil {
		return c.JSON(404, helpers.ErrorJSON("用户不存在"))
	}
	var req struct {
		FullName string `json:"full_name"`
		Role     string `json:"role"`
		IsActive *bool  `json:"is_active"`
		Password string `json:"password"`
		Bio      string `json:"bio"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Role != "" {
		switch req.Role {
		case models.RoleAdmin, models.RoleModerator, models.RoleUser:
			user.Role = req.Role
		default:
			return c.JSON(400, helpers.ErrorJSON("无效的角色"))
		}
	}
	if req.IsActive != nil {
		// goent Save 会跳过带 default 且为零值的字段，is_active=false 无法通过 Save 持久化，
		// 需对 is_active 列执行显式 UPDATE。
		if err := db.User.Update().
			Set(goent.Pair{Key: "is_active", Value: *req.IsActive}).
			Where("username = ?", username).
			Exec(); err != nil {
			return c.JSON(500, helpers.ErrorJSON("更新用户失败"))
		}
		user.IsActive = *req.IsActive
	}
	if req.Password != "" {
		if len(req.Password) < 6 {
			return c.JSON(400, helpers.ErrorJSON("密码至少 6 位"))
		}
		if err := user.SetPassword(req.Password); err != nil {
			return c.JSON(500, helpers.ErrorJSON("密码加密失败"))
		}
	}
	user.UpdatedAt = timeNow()
	if err := db.User.Save().One(user); err != nil {
		return c.JSON(500, helpers.ErrorJSON("更新用户失败"))
	}
	return c.JSON(200, ToUserResponse(user))
}
