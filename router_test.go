package hr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestUnsafeParse(t *testing.T) {
	cases := []string{
		"/",
		"/foo",
		"/foo/bar",
		"/foo/bar/",
	}

	c := make([][]byte, 0)

	for _, v := range cases {
		c = c[:0] // reset c
		unsafeParse(v, &c)
		a := bytes.Split([]byte(v), []byte{'/'})
		if !reflect.DeepEqual(c, a) {
			t.Fatalf("[%s]:\nwant: %v\ngot:  %v", v, a, c)
		}
	}
}

func TestRouter(t *testing.T) {
	r := Default()
	r.Handle("GET", "/", testHandler(0))
	r.Handle("GET", "/:foo", testHandler(1))
	r.Handle("GET", "/:foo/:bar", testHandler(2))
	r.Handle("GET", "/ping/pong", testHandler(3))

	g1 := r.Prefix("/g1")
	g1.Handle("GET", "/foo", testHandler(10))

	g2 := r.Prefix("/g2")
	g2.Handle("GET", "/foo", testHandler(20))

	cases := []struct {
		method  string
		path    string
		code    int
		handler int
	}{
		{"GET", "/", 200, 1},
		{"GET", "/foo", 200, 1},
		{"GET", "/foo/bar", 200, 2},
		{"GET", "/ping/pong", 200, 3},
		{"GET", "/ping/pong/xxx", 200, 0},
		{"GET", "/g1/foo", 200, 10},
		{"GET", "/g2/foo", 200, 20},
		{"POST", "/", 404, -1},
	}

	for _, v := range cases {
		req, _ := http.NewRequest(v.method, v.path, nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)
		if rw.Code != v.code {
			t.Fatalf("[%s] bad status code, want %d got %d", v.path, v.code, rw.Code)
		}
		if v.code == 200 {
			b, _ := io.ReadAll(rw.Body)
			h, _ := strconv.Atoi(string(b))
			if v.handler != h {
				t.Fatalf("[%s] bad response body, want %d got %d", v.path, v.handler, h)
			}
		}
	}
}

func newTestWrapper(n int) Plugin {
	return func(next Handler) Handler {
		return HandlerFunc(func(c *Ctx) error {
			fmt.Println(n)
			next.ServeHTTP(c)
			fmt.Println(n * 10)
			return nil
		})
	}
}

func Example() {
	r := New("/r", newTestWrapper(1))

	h := func(c *Ctx) error {
		fmt.Println(0)
		return nil
	}

	g := r.Prefix("/g", newTestWrapper(2))
	g.GET("/foo", h, newTestWrapper(3))

	req, _ := http.NewRequest("GET", "/r/g/foo", nil)
	rw := httptest.NewRecorder()
	r.ServeHTTP(rw, req)

	// Output:
	// 1
	// 2
	// 3
	// 0
	// 30
	// 20
	// 10
}
