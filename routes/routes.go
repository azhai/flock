package routes

import (
	"github.com/azhai/flock/config"
	"github.com/azhai/flock/handlers"
	"github.com/azhai/flock/middleware"
	"github.com/labstack/echo/v5"
)

func SetupAPIRoutes(e *echo.Echo) {
	api := e.Group(config.APIBaseURL)

	// 公开接口
	api.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})
	api.POST("/auth/register", handlers.Register)
	api.POST("/auth/login", handlers.Login)
	api.GET("/categories", handlers.ListCategories)
	api.GET("/tags", handlers.ListTags)
	api.GET("/topics", handlers.ListTopics, middleware.OptionalAuth())
	api.GET("/topics/:id", handlers.GetTopic, middleware.OptionalAuth())
	api.GET("/topics/:id/posts", handlers.ListPosts, middleware.OptionalAuth())
	api.GET("/users/:username", handlers.GetUser)
	api.GET("/users/:username/topics", handlers.GetUserTopics)
	api.GET("/users/:username/posts", handlers.GetUserPosts)

	// 需登录
	authed := api.Group("", middleware.AuthMiddleware())
	authed.GET("/auth/me", handlers.Me)
	authed.PUT("/user/me", handlers.UpdateProfile)
	authed.GET("/bookmarks", handlers.ListBookmarks)
	authed.POST("/topics/:id/bookmark", handlers.ToggleBookmark)
	authed.POST("/topics/:id/follow", handlers.ToggleFollow)
	authed.POST("/topics", handlers.CreateTopic)
	authed.PUT("/topics/:id", handlers.UpdateTopic)
	authed.DELETE("/topics/:id", handlers.DeleteTopic)
	authed.POST("/topics/:id/posts", handlers.CreatePost)
	authed.POST("/topics/:id/pin", handlers.PinTopic)
	authed.POST("/topics/:id/unpin", handlers.UnpinTopic)
	authed.POST("/topics/:id/close", handlers.CloseTopic)
	authed.POST("/topics/:id/open", handlers.OpenTopic)
	authed.PUT("/posts/:id", handlers.UpdatePost)
	authed.DELETE("/posts/:id", handlers.DeletePost)
	authed.POST("/posts/:id/like", handlers.ToggleLike)
	authed.POST("/posts/:id/vote", handlers.VotePost)
	authed.POST("/posts/:id/report", handlers.CreateReport)

	// 通知
	authed.GET("/notifications", handlers.ListNotifications)
	authed.GET("/notifications/unread", handlers.UnreadNotificationCount)
	authed.POST("/notifications/:id/read", handlers.MarkNotificationRead)
	authed.POST("/notifications/read-all", handlers.MarkAllNotificationsRead)
	authed.GET("/notifications/stream", handlers.StreamNotifications)

	// 仅管理员
	admin := api.Group("", middleware.AuthMiddleware(), middleware.AdminOnly())
	admin.POST("/categories", handlers.CreateCategory)
	admin.PUT("/categories/:id", handlers.UpdateCategory)
	admin.DELETE("/categories/:id", handlers.DeleteCategory)
	admin.GET("/admin/users", handlers.ListUsers)
	admin.PUT("/admin/users/:username", handlers.AdminUpdateUser)
	admin.GET("/reports", handlers.ListReports)
	admin.POST("/reports/:id/resolve", handlers.ResolveReport)
	admin.POST("/reports/:id/dismiss", handlers.DismissReport)
}
