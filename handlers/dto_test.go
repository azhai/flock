package handlers

import (
	"testing"
	"time"

	"github.com/azhai/gouache/models"
)

func TestTrustLevel(t *testing.T) {
	cases := []struct {
		role  string
		posts int64
		lvl   int
		name  string
	}{
		{models.RoleAdmin, 0, 3, "管理员"},
		{models.RoleModerator, 0, 3, "版主"},
		{models.RoleUser, 50, 3, "资深"},
		{models.RoleUser, 100, 3, "资深"},
		{models.RoleUser, 10, 2, "中级"},
		{models.RoleUser, 25, 2, "中级"},
		{models.RoleUser, 1, 1, "初级"},
		{models.RoleUser, 0, 0, "新手"},
		{"", 0, 0, "新手"},
	}
	for _, c := range cases {
		lvl, name := trustLevel(c.role, c.posts)
		if lvl != c.lvl || name != c.name {
			t.Errorf("trustLevel(%q,%d) = %d/%s, want %d/%s", c.role, c.posts, lvl, name, c.lvl, c.name)
		}
	}
}

func TestToUserResponseNilAndFull(t *testing.T) {
	if ToUserResponse(nil) != nil {
		t.Fatal("nil user should be nil response")
	}
	now := time.Now()
	u := &models.User{
		ID: 1, Username: "u", FullName: "F", Avatar: "a", Bio: "b", Website: "w", Location: "l",
		Role: models.RoleAdmin, IsActive: true, CreatedAt: now, PostsCount: 99,
	}
	r := ToUserResponse(u)
	if r == nil || r.Username != "u" || r.Level != 3 || r.LevelName != "管理员" || r.PostsCount != 99 {
		t.Fatalf("ToUserResponse bad: %+v", r)
	}
}

func TestCategoryResponses(t *testing.T) {
	if ToCategoryResponse(nil, 0) != nil {
		t.Fatal("nil cat")
	}
	parent := int64(3)
	cat := &models.Category{ID: 2, Name: "cat", Slug: "cat", Color: "#000", Position: 1, ParentID: &parent}
	r := ToCategoryResponse(cat, 5)
	if r == nil || r.ID != 2 || r.TopicsCount != 5 || r.ParentID == nil || *r.ParentID != 3 {
		t.Fatalf("cat bad: %+v", r)
	}

	// buildCategoryTree：根 + 子，以及悬挂父级
	cats := []*models.Category{
		{ID: 1, Name: "root", Slug: "root"},
		{ID: 2, Name: "child", Slug: "child"},
	}
	// 用非 nil ParentID 指向已存在父
	p1 := int64(1)
	cats[1].ParentID = &p1
	tree := buildCategoryTree(cats, map[int64]int64{1: 2, 2: 0})
	if len(tree) != 1 || tree[0].ID != 1 || len(tree[0].Children) != 1 {
		t.Fatalf("tree bad: %+v", tree)
	}

	// 悬挂父级 → 成为根
	dangling := []*models.Category{{ID: 9, Name: "d", Slug: "d"}}
	dp := int64(88)
	dangling[0].ParentID = &dp
	if tree2 := buildCategoryTree(dangling, nil); len(tree2) != 1 {
		t.Fatalf("dangling tree bad: %+v", tree2)
	}
}

func TestTagResponse(t *testing.T) {
	if ToTagResponse(nil) != nil {
		t.Fatal("nil tag")
	}
	tg := &models.Tag{ID: 3, Name: "go", Slug: "go", TopicsCount: 7}
	r := ToTagResponse(tg)
	if r == nil || r.Name != "go" || r.TopicsCount != 7 {
		t.Fatalf("tag bad: %+v", r)
	}
}

func TestTopicPostResponses(t *testing.T) {
	if ToTopicResponse(nil, 0) != nil {
		t.Fatal("nil topic")
	}
	if ToPostResponse(nil, false) != nil {
		t.Fatal("nil post")
	}
	now := time.Now()
	topic := &models.Topic{
		ID: 1, Title: "t", Slug: "t", User: &models.User{ID: 1, Username: "u"},
		Category:     &models.Category{ID: 2, Name: "c"},
		Tags:         []*models.Tag{{ID: 1, Name: "x"}},
		Participants: []*models.User{{ID: 2, Username: "p"}, {ID: 3, Username: "q"}},
		Pinned:       true, Closed: true, CreatedAt: now, UpdatedAt: now,
		ViewsCount: 4, PostsCount: 2, ReplyCount: 1, LikeCount: 3,
	}
	tr := ToTopicResponse(topic, 9)
	if tr == nil || tr.Category == nil || len(tr.Tags) != 1 || len(tr.Participants) != 2 {
		t.Fatalf("topic resp bad: %+v", tr)
	}
	post := &models.Post{ID: 1, User: &models.User{ID: 1, Username: "u"}, CreatedAt: now, UpdatedAt: now}
	pr := ToPostResponse(post, true)
	if pr == nil || !pr.Liked || pr.User == nil {
		t.Fatalf("post resp bad: %+v", pr)
	}
}

func TestNotificationAndReportResponses(t *testing.T) {
	if ToNotificationResponse(nil) != nil || ToReportResponse(nil) != nil {
		t.Fatal("nil responses")
	}
	now := time.Now()
	n := &models.Notification{ID: 1, Type: models.NotifyReply, Actor: &models.User{ID: 1, Username: "a"}, TopicID: 5, PostID: 6, Content: "c", CreatedAt: now}
	nr := ToNotificationResponse(n)
	if nr == nil || nr.Actor == nil || nr.Type != models.NotifyReply {
		t.Fatalf("notif resp bad: %+v", nr)
	}
	r := &models.Report{ID: 1, Reason: "r", Status: models.ReportPending, User: &models.User{ID: 1}, Post: &models.Post{ID: 2}}
	rr := ToReportResponse(r)
	if rr == nil || rr.User == nil || rr.Post == nil {
		t.Fatalf("report resp bad: %+v", rr)
	}
}

func TestTimeNow(t *testing.T) {
	if timeNow().IsZero() {
		t.Fatal("timeNow zero")
	}
}
