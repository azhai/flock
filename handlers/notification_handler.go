package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/middleware"
	"github.com/azhai/flock/models"
	"github.com/azhai/goent"
	"github.com/labstack/echo/v5"
)

// —— 通知创建（best-effort，不影响主流程）——

func notify(userID, actorID int64, typ string, topicID, postID int64, content string) error {
	if userID == 0 || userID == actorID {
		return nil
	}
	n := &models.Notification{
		UserID:  userID,
		ActorID: actorID,
		Type:    typ,
		TopicID: topicID,
		PostID:  postID,
		Content: content,
	}
	return models.GetDB().Notification.Insert().One(n)
}

// mentionRe 匹配 @用户名（含中文/字母/数字/连字符/下划线）
var mentionRe = regexp.MustCompile("@([\\p{L}\\p{N}_-]+)")

// createReplyNotifications 回复话题 → 通知话题作者 + 所有关注者（不含自己）
func createReplyNotifications(post *models.Post, topic *models.Topic, actorID int64) {
	db := models.GetDB()
	recipients := map[int64]bool{}
	if topic.UserID != 0 && topic.UserID != actorID {
		recipients[topic.UserID] = true
	}
	if fs, _ := db.TopicFollower.Select().
		Filter(goent.Equals(db.TopicFollower.Field("topic_id"), topic.ID)).
		All(); fs != nil {
		for _, f := range fs {
			if f.UserID != actorID {
				recipients[f.UserID] = true
			}
		}
	}
	for uid := range recipients {
		_ = notify(uid, actorID, models.NotifyReply, topic.ID, post.ID, topic.Title)
	}
}

// createMentionNotifications 解析 @提及并通知被提及者
func createMentionNotifications(post *models.Post, actorID int64) {
	names := mentionRe.FindAllStringSubmatch(post.Raw, -1)
	seen := map[string]bool{}
	usernames := make([]string, 0, len(names))
	for _, m := range names {
		name := strings.ToLower(strings.TrimSpace(m[1]))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		usernames = append(usernames, name)
	}
	if len(usernames) == 0 {
		return
	}
	db := models.GetDB()
	users, err := db.User.Select().
		Filter(goent.In(db.User.Field("username"), usernames)).
		All()
	if err != nil {
		return
	}
	topic := &models.Topic{}
	if t, _ := db.Topic.Select().Where("id = ?", post.TopicID).One(); t != nil {
		topic = t
	}
	for _, u := range users {
		_ = notify(u.ID, actorID, models.NotifyMention, post.TopicID, post.ID, topic.Title)
	}
}

// createLikeNotification 收到点赞 → 通知帖子作者
func createLikeNotification(post *models.Post, actorID int64) {
	_ = notify(post.UserID, actorID, models.NotifyLike, post.TopicID, post.ID, "")
}

// createReportNotifications 新举报 → 通知所有版主/管理员
func createReportNotifications(r *models.Report, topicTitle string) {
	db := models.GetDB()
	staff, _ := db.User.Select().
		Where("role IN ('admin', 'moderator')").
		All()
	for _, s := range staff {
		_ = notify(s.ID, r.UserID, models.NotifyReport, r.TopicID, r.PostID, topicTitle)
	}
}

// notificationSnapshot 取某用户未读摘要，用于 SSE 对比
func notificationSnapshot(userID int64) (int64, int64) {
	db := models.GetDB()
	unread, _ := db.Notification.Filter(goent.And(
		goent.Equals(db.Notification.Field("user_id"), userID),
		goent.Equals(db.Notification.Field("read"), false),
	)).Count("id")
	var last int64
	if n, err := db.Notification.Select().
		Where("user_id = ?", userID).
		OrderBy("id DESC").
		Take(1).
		One(); err == nil && n != nil {
		last = n.ID
	}
	return unread, last
}

// ListNotifications 最近通知列表
func ListNotifications(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	db := models.GetDB()
	notifs, err := db.Notification.Select().
		Where("user_id = ?", user.ID).
		OrderBy("id DESC").
		Take(50).
		With("Actor").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载通知失败"))
	}
	resp := make([]*NotificationResponse, 0, len(notifs))
	for _, n := range notifs {
		resp = append(resp, ToNotificationResponse(n))
	}
	return c.JSON(200, resp)
}

// UnreadNotificationCount 未读通知数
func UnreadNotificationCount(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	unread, _ := notificationSnapshot(user.ID)
	return c.JSON(200, map[string]any{"unread": unread})
}

func MarkNotificationRead(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid id"))
	}
	db := models.GetDB()
	err = db.Notification.Update().
		Set(goent.Pair{Key: "read", Value: true}).
		Filter(goent.And(
			goent.Equals(db.Notification.Field("id"), id),
			goent.Equals(db.Notification.Field("user_id"), user.ID),
		)).
		Exec()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	return c.JSON(200, map[string]any{"ok": true})
}

func MarkAllNotificationsRead(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	db := models.GetDB()
	err := db.Notification.Update().
		Set(goent.Pair{Key: "read", Value: true}).
		Filter(goent.Equals(db.Notification.Field("user_id"), user.ID)).
		Exec()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	return c.JSON(200, map[string]any{"ok": true})
}

// StreamNotifications SSE 推送未读数（服务端轮询，兼容多请求无 broker 场景）
func StreamNotifications(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	w := c.Response()
	cons := c.Request().Context()
	w.Header().Set(echo.HeaderContentType, "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)

	send := func(count int64) error {
		if _, err := w.Write([]byte("event: notify\ndata: {\"unread\":" + strconv.FormatInt(count, 10) + "}\n\n")); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	lastUnread, lastID := notificationSnapshot(user.ID)
	if err := send(lastUnread); err != nil {
		return err
	}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-cons.Done():
			return nil
		case <-ticker.C:
			unread, id := notificationSnapshot(user.ID)
			if id != lastID || unread != lastUnread {
				if err := send(unread); err != nil {
					return nil
				}
				lastUnread, lastID = unread, id
			} else if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}
