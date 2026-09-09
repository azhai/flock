//go:generate goent-gen fields --force ./
package models

import (
	"time"

	"github.com/azhai/goent"
	"golang.org/x/crypto/bcrypt"
)

// ForumSchema 论坛所有表。表名默认 snake_case（User→user, PostLike→post_like）。
type ForumSchema struct {
	User          *goent.Table[User]
	Category      *goent.Table[Category]
	Topic         *goent.Table[Topic]
	Post          *goent.Table[Post]
	PostLike      *goent.Table[PostLike]
	Tag           *goent.Table[Tag]
	TopicTag      *goent.Table[TopicTag]
	Bookmark      *goent.Table[Bookmark]
	Notification  *goent.Table[Notification]
	TopicFollower *goent.Table[TopicFollower]
	PostVote      *goent.Table[PostVote]
	Report        *goent.Table[Report]
}

// 用户角色常量
const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleUser      = "user"
)

// User 用户
type User struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`
	UpdatedAt time.Time `goe:"default:current_timestamp"`

	Username string `goe:"unique"`
	Email    string `goe:"unique"`
	Password string
	FullName string
	Avatar   string
	Bio      string
	Website  string
	Location string

	IsActive bool   `goe:"default:true"`
	Role     string `goe:"default:user"` // admin / moderator / user

	// 附加挂载（非列，仅用于统计/信任等级派生）
	TopicsCount int64 `goe:"-"`
	PostsCount  int64 `goe:"-"`
}

func (u *User) SetPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (u *User) IsStaff() bool {
	return u.Role == RoleAdmin || u.Role == RoleModerator
}

// Category 分类，支持嵌套（ParentID 指向自身，空为顶级）
type Category struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`
	UpdatedAt time.Time `goe:"default:current_timestamp"`

	Name        string `goe:"unique"`
	Slug        string `goe:"unique"`
	Description string
	Color       string `goe:"default:'#718096'"`
	Position    int    `goe:"default:0"`
	ParentID    *int64 `goe:"index;m2o"`

	// 预加载挂载字段
	Parent   *Category   `goe:"-"`
	Children []*Category `goe:"o2m,fk:parent_id"`
}

// Topic 话题（Discourse 模型：话题 + 首帖）
type Topic struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`
	UpdatedAt time.Time `goe:"default:current_timestamp"`

	Title      string
	Slug       string `goe:"index"`
	CategoryID *int64 `goe:"index;m2o"`
	UserID     int64  `goe:"index;m2o"`

	Pinned   bool `goe:"default:false;index"`
	PinOrder int  `goe:"default:0"`
	Closed   bool `goe:"default:false"`
	Hidden   bool `goe:"default:false;index"`

	ViewsCount   int        `goe:"default:0"`
	PostsCount   int        `goe:"default:0"`
	ReplyCount   int        `goe:"default:0"`
	LikeCount    int        `goe:"default:0"`
	LastPostID   *int64     `goe:"index"`
	LastPostedAt *time.Time `goe:"index"`

	// 预加载挂载字段
	Category *Category `goe:"-"`
	User     *User     `goe:"-"`
	Posts    []*Post   `goe:"o2m,fk:topic_id"`

	// 附加挂载（非列，仅用于 DTO 组装）
	Tags         []*Tag  `goe:"-"`
	Participants []*User `goe:"-"`
	Bookmarked   bool    `goe:"-"`
	Followed     bool    `goe:"-"`
}

// Post 帖子（话题下的每个楼层；话题首帖的 PostNumber=1）
type Post struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`
	UpdatedAt time.Time `goe:"default:current_timestamp"`

	TopicID int64 `goe:"index;m2o"`
	UserID  int64 `goe:"index;m2o"`

	PostNumber        int    `goe:"index"`
	Raw               string // Markdown 原文
	ReplyToPostNumber int    `goe:"default:0"`

	LikeCount int `goe:"default:0"`

	// 预加载挂载字段
	Topic *Topic      `goe:"-"`
	User  *User       `goe:"-"`
	Likes []*PostLike `goe:"o2m,fk:post_id"`

	// 附加挂载（非列，仅用于 DTO 组装）
	VoteScore int `goe:"-"`
	MyVote    int `goe:"-"`
}

// PostLike 帖子点赞（post_id + user_id 联合唯一）
type PostLike struct {
	ID        int64 `goe:"pk"`
	CreatedAt time.Time

	PostID int64 `goe:"index(unique n:idx_post_like)"`
	UserID int64 `goe:"index(unique n:idx_post_like)"`
}

// Tag 标签（Name/Slug 唯一；TopicsCount 反规范化计数）
type Tag struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`

	Name        string `goe:"unique"`
	Slug        string `goe:"unique"`
	TopicsCount int    `goe:"default:0"`
}

// TopicTag 话题-标签关联（topic_id + tag_id 联合唯一；不加 m2o 以保留唯一索引，手动联表）
type TopicTag struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`

	TopicID int64 `goe:"index(unique n:idx_topic_tag);index"`
	TagID   int64 `goe:"index(unique n:idx_topic_tag);index"`
}

// Bookmark 话题收藏（user_id + topic_id 联合唯一；镜像 PostLike）
type Bookmark struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`

	UserID  int64 `goe:"index(unique n:idx_bookmark);index"`
	TopicID int64 `goe:"index(unique n:idx_bookmark);index"`
}

// 通知类型常量
const (
	NotifyReply   = "reply"
	NotifyMention = "mention"
	NotifyLike    = "like"
	NotifyFollow  = "follow"
	NotifyReport  = "report"
)

// Notification 用户通知（收件人=UserID）
type Notification struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`
	Read      bool      `goe:"default:false;index"`

	UserID  int64  `goe:"index;m2o"` // 收件人
	ActorID int64  `goe:"index;m2o"` // 触发者
	Type    string `goe:"index"`
	TopicID int64  `goe:"index"`
	PostID  int64  `goe:"index"`
	Content string // 摘要（话题标题 / 被赞内容等）

	// 预加载挂载字段
	Actor *User  `goe:"-"`
	Topic *Topic `goe:"-"`
}

// TopicFollower 话题关注者（user_id + topic_id 联合唯一）
type TopicFollower struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`

	UserID  int64 `goe:"index(unique n:idx_topic_follower);index"`
	TopicID int64 `goe:"index(unique n:idx_topic_follower);index"`
}

// PostVote 帖子评分（顶=1 / 踩=-1 / 撤销=0；post+user 联合唯一）
type PostVote struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`

	PostID int64 `goe:"index(unique n:idx_post_vote);index"`
	UserID int64 `goe:"index(unique n:idx_post_vote);index"`
	Vote   int   `goe:"default:0;index"`
}

// 举报状态常量
const (
	ReportPending   = "pending"
	ReportResolved  = "resolved"
	ReportDismissed = "dismissed"
)

// Report 举报（举报人=UserID，目标=PostID）
type Report struct {
	ID        int64     `goe:"pk"`
	CreatedAt time.Time `goe:"default:current_timestamp"`

	UserID  int64 `goe:"index;m2o"` // 举报人
	PostID  int64 `goe:"index;m2o"`
	TopicID int64 `goe:"index"`
	Reason  string
	Status  string `goe:"default:pending;index"`

	// 预加载挂载字段
	User *User `goe:"-"`
	Post *Post `goe:"-"`
}
