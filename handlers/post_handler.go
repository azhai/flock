package handlers

import (
	"strconv"
	"strings"

	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/middleware"
	"github.com/azhai/flock/models"
	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
	"github.com/labstack/echo/v5"
)

type CreatePostRequest struct {
	Raw               string `json:"raw"`
	ReplyToPostNumber int    `json:"reply_to_post_number"`
}

// ListPosts 话题帖子流（按楼号顺序），登录用户附带 liked 标记
func ListPosts(c *echo.Context) error {
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid topic id"))
	}
	db := models.GetDB()
	p := helpers.GetPagination(c)

	total, _ := db.Post.Filter(goent.Equals(db.Post.Field("topic_id"), topicID)).Count("id")
	posts, err := db.Post.Select().
		Where("topic_id = ?", topicID).
		OrderBy("post_number ASC", "id ASC").
		Skip(helpers.GetOffset(p.Page, p.PerPage)).
		Take(p.PerPage).
		With("User").
		All()
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("加载帖子失败"))
	}

	// 批量查询当前用户点赞
	liked := make(map[int64]bool)
	if userID := middleware.GetCurrentUserID(c); userID != 0 && len(posts) > 0 {
		ids := make([]int64, len(posts))
		for i, post := range posts {
			ids[i] = post.ID
		}
		likes, _ := db.PostLike.Select().
			Filter(goent.And(
				goent.Equals(db.PostLike.Field("user_id"), userID),
				goent.In(db.PostLike.Field("post_id"), ids),
			)).
			All()
		for _, like := range likes {
			liked[like.PostID] = true
		}
	}

	resp := make([]*PostResponse, 0, len(posts))
	attachAuthorStats(posts)
	voteScores, myVotes := loadPostVotes(postIDs(posts), middleware.GetCurrentUserID(c))
	for _, post := range posts {
		post.VoteScore = voteScores[post.ID]
		post.MyVote = myVotes[post.ID]
		resp = append(resp, ToPostResponse(post, liked[post.ID]))
	}
	return c.JSON(200, helpers.NewPaginatedResponse(resp, p.Page, p.PerPage, total))
}

// postIDs 取帖子 id 列表
func postIDs(posts []*models.Post) []int64 {
	ids := make([]int64, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	return ids
}

// loadPostVotes 批量加载评分与本人投票：post_id → 总分 / post_id → 我的票
func loadPostVotes(ids []int64, userID int64) (map[int64]int, map[int64]int) {
	scores := map[int64]int{}
	mine := map[int64]int{}
	if len(ids) == 0 {
		return scores, mine
	}
	votes, _ := models.GetDB().PostVote.Select().
		Filter(goent.In(models.GetDB().PostVote.Field("post_id"), ids)).
		All()
	for _, v := range votes {
		scores[v.PostID] += v.Vote
		if userID != 0 && v.UserID == userID {
			mine[v.PostID] = v.Vote
		}
	}
	return scores, mine
}

// attachAuthorStats 逐作者统计发帖量（用于信任等级徽章），页内作者数受 perPage 约束
func attachAuthorStats(posts []*models.Post) {
	db := models.GetDB()
	seen := map[int64]bool{}
	for _, p := range posts {
		if p.User == nil || seen[p.UserID] {
			continue
		}
		seen[p.UserID] = true
		if n, err := db.Post.Filter(goent.Equals(db.Post.Field("user_id"), p.UserID)).Count("id"); err == nil {
			p.User.PostsCount = n
		}
	}
}

// CreatePost 回复话题
func CreatePost(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid topic id"))
	}
	var req CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	if strings.TrimSpace(req.Raw) == "" {
		return c.JSON(400, helpers.ErrorJSON("内容不能为空"))
	}

	db := models.GetDB()
	topic, err := db.Topic.Select().Where("id = ?", topicID).One()
	if err != nil || topic == nil {
		return c.JSON(404, helpers.ErrorJSON("话题不存在"))
	}
	if topic.Closed {
		return c.JSON(403, helpers.ErrorJSON("话题已锁定，无法回复"))
	}

	post := &models.Post{
		TopicID:           topicID,
		UserID:            user.ID,
		PostNumber:        topic.PostsCount + 1,
		Raw:               req.Raw,
		ReplyToPostNumber: req.ReplyToPostNumber,
	}

	err = db.BeginTransaction(func(tx model.Transaction) error {
		if err := db.Post.Insert().OnTransaction(tx).One(post); err != nil {
			return err
		}
		then := timeNow()
		return db.Topic.Update().
			Set(
				goent.Pair{Key: "posts_count", Value: topic.PostsCount + 1},
				goent.Pair{Key: "reply_count", Value: topic.ReplyCount + 1},
				goent.Pair{Key: "last_post_id", Value: post.ID},
				goent.Pair{Key: "last_posted_at", Value: then},
				goent.Pair{Key: "updated_at", Value: then},
			).
			Filter(goent.Equals(db.Topic.Field("id"), topicID)).
			OnTransaction(tx).
			Exec()
	})
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("回复失败"))
	}

	// 通知：话题作者+关注者被回复、被 @ 提及（best-effort）
	createReplyNotifications(post, topic, user.ID)
	createMentionNotifications(post, user.ID)

	created, _ := db.Post.Select().Where("id = ?", post.ID).With("User").One()
	return c.JSON(201, ToPostResponse(created, false))
}

// UpdatePost 编辑帖子（作者或 staff）
func UpdatePost(c *echo.Context) error {
	post, err := findPost(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("帖子不存在"))
	}
	if !canEditPost(c, post) {
		return c.JSON(403, helpers.ErrorJSON("无权限编辑该帖子"))
	}
	var req struct {
		Raw string `json:"raw"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	if strings.TrimSpace(req.Raw) == "" {
		return c.JSON(400, helpers.ErrorJSON("内容不能为空"))
	}

	db := models.GetDB()
	post.Raw = req.Raw
	post.UpdatedAt = timeNow()
	if err := db.Post.Save().One(post); err != nil {
		return c.JSON(500, helpers.ErrorJSON("更新帖子失败"))
	}
	return c.JSON(200, ToPostResponse(post, false))
}

// DeletePost 删除帖子（作者或 staff）；首帖请删除话题
func DeletePost(c *echo.Context) error {
	post, err := findPost(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("帖子不存在"))
	}
	if !canEditPost(c, post) {
		return c.JSON(403, helpers.ErrorJSON("无权限删除该帖子"))
	}
	if post.PostNumber == 1 {
		return c.JSON(400, helpers.ErrorJSON("这是话题首帖，请直接删除话题"))
	}

	db := models.GetDB()
	err = deletePostInternal(db, post)
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("删除帖子失败"))
	}
	return c.JSON(200, map[string]any{"message": "deleted"})
}

// ToggleLike 切换点赞，返回当前状态
func ToggleLike(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	post, err := findPost(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("帖子不存在"))
	}

	db := models.GetDB()
	existing, _ := db.PostLike.Select().
		Where("post_id = ? AND user_id = ?", post.ID, user.ID).
		One()

	var by int
	if existing != nil {
		by = -1
		err = db.BeginTransaction(func(tx model.Transaction) error {
			if err := db.PostLike.Delete().
				Filter(goent.Equals(db.PostLike.Field("id"), existing.ID)).
				OnTransaction(tx).
				Exec(); err != nil {
				return err
			}
			if err := setPostLikeCount(tx, post.ID, post.LikeCount+by); err != nil {
				return err
			}
			return setTopicLikeCount(tx, post.TopicID, by)
		})
	} else {
		by = 1
		err = db.BeginTransaction(func(tx model.Transaction) error {
			like := &models.PostLike{PostID: post.ID, UserID: user.ID}
			if err := db.PostLike.Insert().OnTransaction(tx).One(like); err != nil {
				return err
			}
			if err := setPostLikeCount(tx, post.ID, post.LikeCount+by); err != nil {
				return err
			}
			return setTopicLikeCount(tx, post.TopicID, by)
		})
	}
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	if by > 0 {
		createLikeNotification(post, user.ID)
	}
	liked := by > 0
	return c.JSON(200, map[string]any{"liked": liked, "like_count": post.LikeCount + by})
}

func setPostLikeCount(tx model.Transaction, postID int64, val int) error {
	db := models.GetDB()
	return db.Post.Update().
		Set(goent.Pair{Key: "like_count", Value: val}).
		Filter(goent.Equals(db.Post.Field("id"), postID)).
		OnTransaction(tx).
		Exec()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func setTopicLikeCount(tx model.Transaction, topicID int64, by int) error {
	db := models.GetDB()
	topic, err := db.Topic.Select().Where("id = ?", topicID).OnTransaction(tx).One()
	if err != nil || topic == nil {
		return err
	}
	return db.Topic.Update().
		Set(goent.Pair{Key: "like_count", Value: maxInt(topic.LikeCount+by, 0)}).
		Filter(goent.Equals(db.Topic.Field("id"), topicID)).
		OnTransaction(tx).
		Exec()
}
func findPost(c *echo.Context) (*models.Post, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}
	return models.GetDB().Post.Select().Where("id = ?", id).One()
}

func canEditPost(c *echo.Context, post *models.Post) bool {
	if middleware.IsStaff(c) {
		return true
	}
	return middleware.GetCurrentUserID(c) == post.UserID
}
