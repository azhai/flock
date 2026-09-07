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

// CreateReport 举报帖子（同一用户对同一帖子仅保留一条 pending 举报）
func CreateReport(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	post, err := findPost(c)
	if err != nil || post == nil {
		return c.JSON(404, helpers.ErrorJSON("帖子不存在"))
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	if req.Reason == "" {
		return c.JSON(400, helpers.ErrorJSON("请填写举报理由"))
	}

	db := models.GetDB()
	// 已有同人同帖 pending 举报则更新理由，不做重复
	existing, _ := db.Report.Select().
		Where("user_id = ? AND post_id = ? AND status = ?", user.ID, post.ID, models.ReportPending).
		One()
	if existing != nil {
		existing.Reason = req.Reason
		existing.CreatedAt = timeNow()
		if err := db.Report.Save().One(existing); err != nil {
			return c.JSON(500, helpers.ErrorJSON("提交失败"))
		}
		return c.JSON(200, map[string]any{"id": existing.ID, "status": existing.Status})
	}

	r := &models.Report{
		UserID:  user.ID,
		PostID:  post.ID,
		TopicID: post.TopicID,
		Reason:  req.Reason,
		Status:  models.ReportPending,
	}
	if err := db.Report.Insert().One(r); err != nil {
		return c.JSON(500, helpers.ErrorJSON("提交失败"))
	}
	topicTitle := ""
	if t, _ := db.Topic.Select().Where("id = ?", post.TopicID).One(); t != nil {
		topicTitle = t.Title
	}
	createReportNotifications(r, topicTitle)
	return c.JSON(201, map[string]any{"id": r.ID, "status": r.Status})
}

// ListReports 举报流（staff），可按 status 过滤
func ListReports(c *echo.Context) error {
	db := models.GetDB()
	p := helpers.GetPagination(c)

	query := db.Report.Select()
	if st := c.QueryParam("status"); st != "" {
		query = query.Where("status = ?", st)
	}
	total, _ := query.Count("id")
	reports, err := query.
		OrderBy("id DESC").
		Skip(helpers.GetOffset(p.Page, p.PerPage)).
		Take(p.PerPage).
		With("User", "Post").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载举报失败"))
	}

	resp := make([]*ReportResponse, 0, len(reports))
	for _, r := range reports {
		resp = append(resp, ToReportResponse(r))
	}
	return c.JSON(200, helpers.NewPaginatedResponse(resp, p.Page, p.PerPage, total))
}

// ResolveReport 处理举报（通过）
func ResolveReport(c *echo.Context) error {
	return setReportStatus(c, models.ReportResolved)
}

// DismissReport 驳回举报（忽略）
func DismissReport(c *echo.Context) error {
	return setReportStatus(c, models.ReportDismissed)
}

func setReportStatus(c *echo.Context, status string) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid report id"))
	}
	db := models.GetDB()
	report, err := db.Report.Select().Where("id = ?", id).With("User", "Post").One()
	if err != nil || report == nil {
		return c.JSON(404, helpers.ErrorJSON("举报不存在"))
	}
	if report.Status == status {
		return c.JSON(200, ToReportResponse(report))
	}
	report.Status = status
	if err := db.Report.Save().One(report); err != nil {
		return c.JSON(500, helpers.ErrorJSON("更新失败"))
	}
	// resolved 且目标是违规帖：可选地删除帖子
	if status == models.ReportResolved {
		if p, _ := db.Post.Select().Where("id = ?", report.PostID).One(); p != nil && p.PostNumber != 1 {
			_ = deletePostInternal(db, p)
		}
	}
	return c.JSON(200, ToReportResponse(report))
}

// deletePostInternal 在事务里删除帖子并清理关联（含评分/点赞/举报/通知），供举报处理复用
func deletePostInternal(dba *models.Database, post *models.Post) error {
	return dba.BeginTransaction(func(tx model.Transaction) error {
		if err := dba.PostLike.Delete().
			Filter(goent.Equals(dba.PostLike.Field("post_id"), post.ID)).
			OnTransaction(tx).Exec(); err != nil {
			return err
		}
		if err := dba.PostVote.Delete().
			Filter(goent.Equals(dba.PostVote.Field("post_id"), post.ID)).
			OnTransaction(tx).Exec(); err != nil {
			return err
		}
		if err := dba.Report.Delete().
			Filter(goent.Equals(dba.Report.Field("post_id"), post.ID)).
			OnTransaction(tx).Exec(); err != nil {
			return err
		}
		if err := dba.Notification.Delete().
			Filter(goent.Equals(dba.Notification.Field("post_id"), post.ID)).
			OnTransaction(tx).Exec(); err != nil {
			return err
		}
		if err := dba.Post.Delete().
			Filter(goent.Equals(dba.Post.Field("id"), post.ID)).
			OnTransaction(tx).Exec(); err != nil {
			return err
		}
		// 重算话题计数与最后回复
		topic, err := dba.Topic.Select().Where("id = ?", post.TopicID).OnTransaction(tx).One()
		if err != nil || topic == nil {
			return nil
		}
		last, _ := dba.Post.Select().
			Where("topic_id = ?", post.TopicID).
			OrderBy("post_number DESC").
			OnTransaction(tx).One()
		then := timeNow()
		pairs := []goent.Pair{
			{Key: "posts_count", Value: maxInt(topic.PostsCount-1, 0)},
			{Key: "reply_count", Value: maxInt(topic.ReplyCount-1, 0)},
			{Key: "updated_at", Value: then},
		}
		if last != nil {
			pairs = append(pairs,
				goent.Pair{Key: "last_post_id", Value: last.ID},
				goent.Pair{Key: "last_posted_at", Value: last.CreatedAt},
			)
		}
		return dba.Topic.Update().
			Set(pairs...).
			Filter(goent.Equals(dba.Topic.Field("id"), post.TopicID)).
			OnTransaction(tx).Exec()
	})
}
