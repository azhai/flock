package helpers

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v5"
)

func echoCtx(raw string) *echo.Context {
	e := echo.New()
	req := httptest.NewRequest("GET", "/?"+raw, nil)
	return e.NewContext(req, httptest.NewRecorder())
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"":               "topic",
		"   ":            "topic",
		"Hello World":    "hello-world",
		"Foo  Bar  Baz":  "foo-bar-baz",
		"--x--":          "x",
		"中文标题测试":         "中文标题测试",
		"Go & Rust":      "go-rust",
		"Hello__World!!": "hello-world",
		"123":            "123",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetPagination(t *testing.T) {
	if p := GetPagination(echoCtx("")); p.Page != 1 || p.PerPage != 20 {
		t.Errorf("defaults wrong: %+v", p)
	}
	if p := GetPagination(echoCtx("page=0&per_page=999")); p.Page != 1 || p.PerPage != 20 {
		t.Errorf("clamp wrong: %+v", p)
	}
	if p := GetPagination(echoCtx("page=3&per_page=50")); p.Page != 3 || p.PerPage != 50 {
		t.Errorf("custom wrong: %+v", p)
	}
	if p := GetPagination(echoCtx("page=-2&per_page=-5")); p.Page != 1 || p.PerPage != 20 {
		t.Errorf("negative wrong: %+v", p)
	}
}

func TestGetOffset(t *testing.T) {
	if got := GetOffset(1, 20); got != 0 {
		t.Errorf("offset 1,20 = %d", got)
	}
	if got := GetOffset(3, 10); got != 20 {
		t.Errorf("offset 3,10 = %d", got)
	}
}

func TestResponses(t *testing.T) {
	if r := NewResponse("x"); r.Data != "x" || r.Page != 0 {
		t.Errorf("NewResponse: %+v", r)
	}
	r := NewPaginatedResponse("d", 2, 5, 9)
	if !reflect.DeepEqual(r, &Response{Data: "d", Page: 2, PerPage: 5, Total: 9}) {
		t.Errorf("NewPaginatedResponse: %+v", r)
	}
	if e := ErrorJSON("boom"); e["error"] != "boom" {
		t.Errorf("ErrorJSON: %+v", e)
	}
}
