package handlers

import (
	"net/mail"
	"strings"

	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/middleware"
	"github.com/azhai/flock/models"
	"github.com/labstack/echo/v5"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register 用户注册
func Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if len(req.Username) < 3 || len(req.Username) > 32 {
		return c.JSON(400, helpers.ErrorJSON("用户名长度须为 3-32 个字符"))
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.JSON(400, helpers.ErrorJSON("邮箱格式不正确"))
	}
	if len(req.Password) < 6 {
		return c.JSON(400, helpers.ErrorJSON("密码至少 6 位"))
	}

	db := models.GetDB()
	if u, _ := db.User.Select().Where("username = ?", req.Username).One(); u != nil {
		return c.JSON(409, helpers.ErrorJSON("用户名已存在"))
	}
	if u, _ := db.User.Select().Where("email = ?", req.Email).One(); u != nil {
		return c.JSON(409, helpers.ErrorJSON("邮箱已被注册"))
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		Role:     models.RoleUser,
		IsActive: true,
	}
	if err := user.SetPassword(req.Password); err != nil {
		return c.JSON(500, helpers.ErrorJSON("密码加密失败"))
	}
	if err := db.User.Insert().One(user); err != nil {
		return c.JSON(500, helpers.ErrorJSON("创建用户失败"))
	}

	token, err := middleware.GenerateToken(user)
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("生成令牌失败"))
	}
	return c.JSON(201, map[string]any{"user": ToUserResponse(user), "token": token})
}

// Login 用户登录
func Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}

	db := models.GetDB()
	user, _ := db.User.Select().Where("username = ? OR email = ?", req.Username, req.Username).One()
	if user == nil || !user.CheckPassword(req.Password) {
		return c.JSON(401, helpers.ErrorJSON("用户名或密码错误"))
	}
	if !user.IsActive {
		return c.JSON(403, helpers.ErrorJSON("账号已被禁用"))
	}

	token, err := middleware.GenerateToken(user)
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("生成令牌失败"))
	}
	return c.JSON(200, map[string]any{"user": ToUserResponse(user), "token": token})
}

// Me 获取当前登录用户
func Me(c *echo.Context) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil || user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	return c.JSON(200, ToUserResponse(user))
}
