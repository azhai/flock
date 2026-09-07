package config

const (
	APIVersion = "v1"
	APIBaseURL = "/api/" + APIVersion
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// 话题列表排序方式
const (
	SortLatest = "latest" // 最近发帖
	SortTop    = "top"    // 回复最多
	SortLikes  = "likes"  // 获赞最多
	SortHot    = "hot"    // 最近活跃
)
