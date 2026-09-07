package handlers

import (
	"testing"
	"time"

	"github.com/azhai/gouache/models"
)

// 覆盖 dto.go 中纯函数的所有分支（nil 早退 + trustLevel 全档位 + 组装循环）。

func TestDTOToUserNil(t *testing.T) {
	if ToUserResponse(nil) != nil {
		t.Fatal("ToUserResponse(nil) should be nil")
	}
}

func TestDTOToCategoryNil(t *testing.T) {
	if ToCategoryResponse(nil, 0) != nil {
		t.Fatal("ToCategoryResponse(nil) should be nil")
	}
}

func TestDTOToTagNil(t *testing.T) {
	if ToTagResponse(nil) != nil {
		t.Fatal("ToTagResponse(nil) should be nil")
	}
}

func TestDTOToPostNil(t *testing.T) {
	if ToPostResponse(nil, false) != nil {
		t.Fatal("ToPostResponse(nil) should be nil")
	}
}

func TestDTOToNotificationNil(t *testing.T) {
	if ToNotificationResponse(nil) != nil {
		t.Fatal("ToNotificationResponse(nil) should be nil")
	}
}

func TestDTOToReportNil(t *testing.T) {
	if ToReportResponse(nil) != nil {
		t.Fatal("ToReportResponse(nil) should be nil")
	}
}

func TestDTOToTopicNil(t *testing.T) {
	if ToTopicResponse(nil, 0) != nil {
		t.Fatal("ToTopicResponse(nil) should be nil")
	}
}

func TestTrustLevelAllBranches(t *testing.T) {
	cases := []struct {
		role      string
		posts     int64
		wantLevel int
		wantName  string
	}{
		{models.RoleAdmin, 0, 3, "管理员"},
		{models.RoleModerator, 0, 3, "版主"},
		{"user", 50, 3, "资深"},
		{"user", 10, 2, "中级"},
		{"user", 1, 1, "初级"},
		{"user", 0, 0, "新手"},
	}
	for _, c := range cases {
		lvl, name := trustLevel(c.role, c.posts)
		if lvl != c.wantLevel || name != c.wantName {
			t.Errorf("trustLevel(%q,%d)=%d,%s want %d,%s", c.role, c.posts, lvl, name, c.wantLevel, c.wantName)
		}
	}
}

func TestDTOToUserWithFields(t *testing.T) {
	u := &models.User{
		ID: 1, Username: "alice", FullName: "A", Avatar: "av", Bio: "b",
		Website: "w", Location: "L", Role: models.RoleUser, IsActive: true,
		CreatedAt: time.Now(), PostsCount: 5,
	}
	r := ToUserResponse(u)
	if r == nil || r.Username != "alice" || r.PostsCount != 5 || r.Level != 1 || r.LevelName != "初级" {
		t.Fatalf("bad ToUserResponse: %+v", r)
	}
}

func TestDTOToTopicWithAssembled(t *testing.T) {
	now := time.Now()
	cat := &models.Category{ID: 7, Name: "tech", Slug: "tech"}
	tag := &models.Tag{ID: 9, Name: "Go", Slug: "go", TopicsCount: 3}
	participant := &models.User{ID: 2, Username: "bob", Role: models.RoleUser}
	lastID := int64(5)
	top := &models.Topic{
		ID: 1, Title: "t", Slug: "t", Category: cat,
		User: &models.User{ID: 1, Username: "alice"},
		Tags: []*models.Tag{tag}, Participants: []*models.User{participant},
		Pinned: true, Closed: false, Bookmarked: true, Followed: true,
		ViewsCount: 1, PostsCount: 2, ReplyCount: 1, LikeCount: 3,
		LastPostedAt: &now, CreatedAt: now, UpdatedAt: now, LastPostID: &lastID,
	}
	r := ToTopicResponse(top, 9)
	if r == nil || r.Category == nil || r.User == nil || len(r.Tags) != 1 || len(r.Participants) != 1 {
		t.Fatalf("bad ToTopicResponse: %+v", r)
	}
	if r.Category.TopicsCount != 9 || r.Participants[0].Username != "bob" {
		t.Fatalf("assembly mismatch: %+v", r)
	}
}

func TestDTOToPostAndReport(t *testing.T) {
	p := &models.Post{
		ID: 1, TopicID: 2, PostNumber: 3,
		User: &models.User{ID: 1, Username: "a"},
		Raw:  "x", ReplyToPostNumber: 0, LikeCount: 4,
		VoteScore: 2, MyVote: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	pr := ToPostResponse(p, true)
	if pr == nil || !pr.Liked || pr.Score != 2 || pr.MyVote != 1 || pr.User == nil {
		t.Fatalf("bad ToPostResponse: %+v", pr)
	}
	r := &models.Report{
		ID: 9, Reason: "ad", Status: models.ReportPending, Post: p,
		User: &models.User{ID: 3, Username: "c"}, TopicID: 2, CreatedAt: time.Now(),
	}
	rr := ToReportResponse(r)
	if rr == nil || rr.Post == nil || rr.User == nil || rr.Status != models.ReportPending {
		t.Fatalf("bad ToReportResponse: %+v", rr)
	}
	n := &models.Notification{
		ID: 1, Type: "reply", Actor: &models.User{ID: 1, Username: "a"},
		TopicID: 2, PostID: 3, Content: "hi", Read: false, CreatedAt: time.Now(),
	}
	nr := ToNotificationResponse(n)
	if nr == nil || nr.Actor == nil || nr.Content != "hi" {
		t.Fatalf("bad ToNotificationResponse: %+v", nr)
	}
}
