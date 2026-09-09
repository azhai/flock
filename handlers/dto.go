package handlers

import (
	"time"

	"github.com/azhai/flock/models"
)

func timeNow() time.Time { return time.Now() }

type UserResponse struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	FullName    string    `json:"full_name"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Website     string    `json:"website"`
	Location    string    `json:"location"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	TopicsCount int64     `json:"topics_count,omitempty"`
	PostsCount  int64     `json:"posts_count,omitempty"`
	Level       int       `json:"level"`
	LevelName   string    `json:"level_name"`
}

// trustLevel 依据发帖量派生简易信任等级（staff 恒为最高）。
func trustLevel(role string, posts int64) (int, string) {
	if role == models.RoleAdmin {
		return 3, "管理员"
	}
	if role == models.RoleModerator {
		return 3, "版主"
	}
	switch {
	case posts >= 50:
		return 3, "资深"
	case posts >= 10:
		return 2, "中级"
	case posts >= 1:
		return 1, "初级"
	default:
		return 0, "新手"
	}
}

func ToUserResponse(u *models.User) *UserResponse {
	if u == nil {
		return nil
	}
	level, name := trustLevel(u.Role, u.PostsCount)
	return &UserResponse{
		ID:         u.ID,
		Username:   u.Username,
		FullName:   u.FullName,
		Avatar:     u.Avatar,
		Bio:        u.Bio,
		Website:    u.Website,
		Location:   u.Location,
		Role:       u.Role,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
		PostsCount: u.PostsCount,
		Level:      level,
		LevelName:  name,
	}
}

type CategoryResponse struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	Description string              `json:"description"`
	Color       string              `json:"color"`
	Position    int                 `json:"position"`
	ParentID    *int64              `json:"parent_id"`
	TopicsCount int64               `json:"topics_count"`
	Children    []*CategoryResponse `json:"children,omitempty"`
}

func ToCategoryResponse(cat *models.Category, topicsCount int64) *CategoryResponse {
	if cat == nil {
		return nil
	}
	return &CategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Slug:        cat.Slug,
		Description: cat.Description,
		Color:       cat.Color,
		Position:    cat.Position,
		ParentID:    cat.ParentID,
		TopicsCount: topicsCount,
	}
}

type TopicResponse struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Slug         string            `json:"slug"`
	CategoryID   *int64            `json:"category_id"`
	Category     *CategoryResponse `json:"category,omitempty"`
	User         *UserResponse     `json:"user"`
	Tags         []*TagResponse    `json:"tags,omitempty"`
	Participants []*UserResponse   `json:"participants,omitempty"`
	Pinned       bool              `json:"pinned"`
	Closed       bool              `json:"closed"`
	Hidden       bool              `json:"hidden"`
	Bookmarked   bool              `json:"bookmarked"`
	Followed     bool              `json:"followed"`
	ViewsCount   int               `json:"views_count"`
	PostsCount   int               `json:"posts_count"`
	ReplyCount   int               `json:"reply_count"`
	LikeCount    int               `json:"like_count"`
	LastPostedAt *time.Time        `json:"last_posted_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func ToTopicResponse(t *models.Topic, catTopicsCount int64) *TopicResponse {
	if t == nil {
		return nil
	}
	resp := &TopicResponse{
		ID:           t.ID,
		Title:        t.Title,
		Slug:         t.Slug,
		CategoryID:   t.CategoryID,
		User:         ToUserResponse(t.User),
		Pinned:       t.Pinned,
		Closed:       t.Closed,
		Hidden:       t.Hidden,
		Bookmarked:   t.Bookmarked,
		Followed:     t.Followed,
		ViewsCount:   t.ViewsCount,
		PostsCount:   t.PostsCount,
		ReplyCount:   t.ReplyCount,
		LikeCount:    t.LikeCount,
		LastPostedAt: t.LastPostedAt,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	if t.Category != nil {
		resp.Category = ToCategoryResponse(t.Category, catTopicsCount)
	}
	for _, tag := range t.Tags {
		resp.Tags = append(resp.Tags, ToTagResponse(tag))
	}
	for _, u := range t.Participants {
		resp.Participants = append(resp.Participants, ToUserResponse(u))
	}
	return resp
}

type TagResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	TopicsCount int    `json:"topics_count"`
}

func ToTagResponse(t *models.Tag) *TagResponse {
	if t == nil {
		return nil
	}
	return &TagResponse{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		TopicsCount: t.TopicsCount,
	}
}

type PostResponse struct {
	ID                int64         `json:"id"`
	TopicID           int64         `json:"topic_id"`
	PostNumber        int           `json:"post_number"`
	User              *UserResponse `json:"user"`
	Raw               string        `json:"raw"`
	ReplyToPostNumber int           `json:"reply_to_post_number"`
	LikeCount         int           `json:"like_count"`
	Liked             bool          `json:"liked"`
	Score             int           `json:"score"`
	MyVote            int           `json:"my_vote"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

func ToPostResponse(p *models.Post, liked bool) *PostResponse {
	if p == nil {
		return nil
	}
	return &PostResponse{
		ID:                p.ID,
		TopicID:           p.TopicID,
		PostNumber:        p.PostNumber,
		User:              ToUserResponse(p.User),
		Raw:               p.Raw,
		ReplyToPostNumber: p.ReplyToPostNumber,
		LikeCount:         p.LikeCount,
		Liked:             liked,
		Score:             p.VoteScore,
		MyVote:            p.MyVote,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

type NotificationResponse struct {
	ID        int64         `json:"id"`
	Type      string        `json:"type"`
	Actor     *UserResponse `json:"actor,omitempty"`
	TopicID   int64         `json:"topic_id"`
	PostID    int64         `json:"post_id"`
	Content   string        `json:"content"`
	Read      bool          `json:"read"`
	CreatedAt time.Time     `json:"created_at"`
}

func ToNotificationResponse(n *models.Notification) *NotificationResponse {
	if n == nil {
		return nil
	}
	return &NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Actor:     ToUserResponse(n.Actor),
		TopicID:   n.TopicID,
		PostID:    n.PostID,
		Content:   n.Content,
		Read:      n.Read,
		CreatedAt: n.CreatedAt,
	}
}

type ReportResponse struct {
	ID        int64         `json:"id"`
	Reason    string        `json:"reason"`
	Status    string        `json:"status"`
	User      *UserResponse `json:"user"`
	Post      *PostResponse `json:"post"`
	TopicID   int64         `json:"topic_id"`
	CreatedAt time.Time     `json:"created_at"`
}

func ToReportResponse(r *models.Report) *ReportResponse {
	if r == nil {
		return nil
	}
	return &ReportResponse{
		ID:        r.ID,
		Reason:    r.Reason,
		Status:    r.Status,
		User:      ToUserResponse(r.User),
		Post:      ToPostResponse(r.Post, false),
		TopicID:   r.TopicID,
		CreatedAt: r.CreatedAt,
	}
}
