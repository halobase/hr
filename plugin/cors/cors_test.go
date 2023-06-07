package cors

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tekqer/hr"
)

func TestCorsPreflightDefault(t *testing.T) {
	opts := Options{}
	r := hr.Default(Cors(opts))
	r.GET("/", func(c *hr.Ctx) error {
		return nil
	})

	req, _ := http.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://tekq.cn")
	rw := httptest.NewRecorder()
	r.ServeHTTP(rw, req)

	cases := []struct {
		Key   string
		Value string
	}{
		{"Vary", "Origin"},
		{"Access-Control-Allow-Origin", "https://tekq.cn"},
		{"Access-Control-Allow-Method", strings.Join(opts.AllowMethods, ",")},
		{"Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ",")},
	}
	for _, h := range cases {
		if !strings.Contains(rw.Header().Get(h.Key), h.Value) {
			t.Fatalf("response header %s has no value %s", h.Key, h.Value)
		}
	}
}

func TestCorsPreflightCustom(t *testing.T) {
	opts := Options{
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"X-Custom-Key"},
		AllowCredentials: true,
		MaxAge:           time.Second * 10,
	}
	r := hr.Default(Cors(opts))
	r.GET("/", func(c *hr.Ctx) error {
		return nil
	})

	req, _ := http.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://tekq.cn")
	rw := httptest.NewRecorder()
	r.ServeHTTP(rw, req)

	cases := []struct {
		Key   string
		Value string
	}{
		{"Vary", "Origin"},
		{"Access-Control-Allow-Origin", "https://tekq.cn"},
		{"Access-Control-Allow-Method", strings.Join(opts.AllowMethods, ",")},
		{"Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ",")},
		{"Access-Control-Allow-Credentials", "true"},
		{"Access-Control-Max-Age", strconv.Itoa(int(opts.MaxAge / time.Second))},
	}
	for _, h := range cases {
		if !strings.Contains(rw.Header().Get(h.Key), h.Value) {
			t.Fatalf("response header %s has no value %s", h.Key, h.Value)
		}
	}
}

func TestCorsWithoutOrigin(t *testing.T) {
	opts := Options{}
	r := hr.Default(Cors(opts))
	r.GET("/", func(c *hr.Ctx) error {
		return nil
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	r.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("bad status code %d (expected 200)", rw.Code)
	}
}

func TestCorsCrossOrigin(t *testing.T) {
	opts := Options{
		ExposeHeaders: []string{"X-Custom-Key"},
	}
	r := hr.Default(Cors(opts))
	r.GET("/", func(c *hr.Ctx) error {
		return nil
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://tekq.cn")
	rw := httptest.NewRecorder()
	r.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("bad status code %d (expected 200)", rw.Code)
	}

	cases := []struct {
		Key   string
		Value string
	}{
		{"Vary", "Origin"},
		{"Access-Control-Allow-Origin", "https://tekq.cn"},
		{"Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ",")},
	}
	for _, h := range cases {
		if !strings.Contains(rw.Header().Get(h.Key), h.Value) {
			t.Fatalf("response header %s has no value %s", h.Key, h.Value)
		}
	}
}
