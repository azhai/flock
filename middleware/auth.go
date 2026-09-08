package middleware

import (
	"strings"
	"time"

	"github.com/azhai/flock/config"
	"github.com/azhai/flock/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.GetTokenExpiry()) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}

func parseToken(c *echo.Context) *Claims {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		return []byte(config.GetJWTSecret()), nil
	})
	if err != nil || !token.Valid {
		return nil
	}
	return claims
}

func setClaims(c *echo.Context, claims *Claims) {
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)
}

// AuthMiddleware 强制登录：无有效 token 返回 401。
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims := parseToken(c)
			if claims == nil {
				return c.JSON(401, map[string]any{"error": "unauthorized"})
			}
			setClaims(c, claims)
			return next(c)
		}
	}
}

// OptionalAuth 可选登录：有 token 则注入用户信息，否则以游客身份继续。
func OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if claims := parseToken(c); claims != nil {
				setClaims(c, claims)
			}
			return next(c)
		}
	}
}

// AdminOnly 仅 admin。
func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if GetCurrentUserRole(c) != models.RoleAdmin {
				return c.JSON(403, map[string]any{"error": "admin access required"})
			}
			return next(c)
		}
	}
}

// StaffOnly 管理员或版主。
func StaffOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role := GetCurrentUserRole(c)
			if role != models.RoleAdmin && role != models.RoleModerator {
				return c.JSON(403, map[string]any{"error": "staff access required"})
			}
			return next(c)
		}
	}
}

func GetCurrentUserID(c *echo.Context) int64 {
	if v := c.Get("user_id"); v != nil {
		return v.(int64)
	}
	return 0
}

func GetCurrentUsername(c *echo.Context) string {
	if v := c.Get("username"); v != nil {
		return v.(string)
	}
	return ""
}

func GetCurrentUserRole(c *echo.Context) string {
	if v := c.Get("role"); v != nil {
		return v.(string)
	}
	return ""
}

func IsStaff(c *echo.Context) bool {
	role := GetCurrentUserRole(c)
	return role == models.RoleAdmin || role == models.RoleModerator
}

// GetCurrentUser 从数据库加载当前登录用户，未登录返回 nil。
func GetCurrentUser(c *echo.Context) (*models.User, error) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		return nil, nil
	}
	db := models.GetDB()
	return db.User.Select().Where("id = ?", userID).One()
}
