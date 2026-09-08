package helpers

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/azhai/flock/config"
	"github.com/labstack/echo/v5"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page    int
	PerPage int
}

func GetPagination(c *echo.Context) PaginationParams {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = config.DefaultPage
	}
	if perPage < 1 || perPage > config.MaxPerPage {
		perPage = config.DefaultPerPage
	}
	return PaginationParams{Page: page, PerPage: perPage}
}

func GetOffset(page, perPage int) int {
	return (page - 1) * perPage
}

// Response 统一 API 响应
type Response struct {
	Data    any   `json:"data"`
	Page    int   `json:"page,omitempty"`
	PerPage int   `json:"per_page,omitempty"`
	Total   int64 `json:"total,omitempty"`
}

func NewResponse(data any) *Response {
	return &Response{Data: data}
}

func NewPaginatedResponse(data any, page, perPage int, total int64) *Response {
	return &Response{Data: data, Page: page, PerPage: perPage, Total: total}
}

// ErrorJSON 统一错误响应体
func ErrorJSON(message string) map[string]any {
	return map[string]any{"error": message}
}

// Slugify 将标题转为 URL 友好的 slug；非 ASCII（如中文）原样保留。
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "topic"
	}
	return result
}
