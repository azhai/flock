package handlers

import (
	"github.com/azhai/flock/helpers"
	"github.com/azhai/flock/middleware"
	"github.com/azhai/flock/models"
	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
	"github.com/labstack/echo/v5"
)

// VoteRequest 顶/踩请求
type VoteRequest struct {
	Vote int `json:"vote"` // -1 踩 / 0 撤销 / 1 顶
}

// VotePost 对帖子顶/踩，返回新评分与我的票
func VotePost(c *echo.Context) error {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return c.JSON(401, helpers.ErrorJSON("unauthorized"))
	}
	post, err := findPost(c)
	if err != nil {
		return c.JSON(404, helpers.ErrorJSON("帖子不存在"))
	}
	var req VoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, helpers.ErrorJSON("invalid request body"))
	}
	if req.Vote < -1 || req.Vote > 1 {
		return c.JSON(400, helpers.ErrorJSON("非法投票值"))
	}

	db := models.GetDB()
	existing, _ := db.PostVote.Select().
		Where("post_id = ? AND user_id = ?", post.ID, user.ID).
		One()

	myVote := 0
	removed := false
	// delta 为 like_count 净变化：顶 +1 / 踩 -1 / 撤销回退原票
	delta := 0
	err = db.BeginTransaction(func(tx model.Transaction) error {
		// 撤销逻辑：请求 0，或与当前票相同（再点一次取消）
		if req.Vote == 0 || (existing != nil && req.Vote == existing.Vote) {
			if existing != nil {
				removed = true
				delta = -existing.Vote
				if err := db.PostVote.Delete().
					Filter(goent.Equals(db.PostVote.Field("id"), existing.ID)).
					OnTransaction(tx).
					Exec(); err != nil {
					return err
				}
			} else {
				return nil
			}
		} else if existing == nil {
			delta = req.Vote
			v := &models.PostVote{PostID: post.ID, UserID: user.ID, Vote: req.Vote}
			if err := db.PostVote.Insert().OnTransaction(tx).One(v); err != nil {
				return err
			}
		} else {
			delta = req.Vote - existing.Vote
			existing.Vote = req.Vote
			if err := db.PostVote.Save().OnTransaction(tx).One(existing); err != nil {
				return err
			}
		}
		// 同步当前获赞数（like_count），保证顶/踩后显示与排序实时一致
		if delta == 0 {
			return nil
		}
		if err := setPostLikeCount(tx, post.ID, post.LikeCount+delta); err != nil {
			return err
		}
		return setTopicLikeCount(tx, post.TopicID, delta)
	})
	if err != nil {
		return c.JSON(500, helpers.ErrorJSON("操作失败"))
	}
	if !removed {
		myVote = req.Vote
	}

	score := 0
	if votes, _ := db.PostVote.Select().
		Where("post_id = ?", post.ID).All(); votes != nil {
		for _, v := range votes {
			score += v.Vote
		}
	}
	return c.JSON(200, map[string]any{"score": score, "my_vote": myVote, "like_count": post.LikeCount + delta})
}
