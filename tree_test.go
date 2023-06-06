package hr

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

var testAlloc = func() Vars {
	return make(Vars, 0, 10)
}

type testHandler int

func (h testHandler) ServeHTTP(c *Ctx) error {
	fmt.Fprintf(c, "%d", h)
	return nil
}

type testCase struct {
	path    string
	vars    Vars
	handler testHandler
	_404    bool
}

func newTestTree(routes []string) *node {
	root := new(node)
	for i, route := range routes {
		chunks := bytes.Split([]byte(route), []byte{'/'})
		root.insert(chunks, 0, testHandler(i))
	}
	return root
}

func TestLookup_Static_Wildcard(t *testing.T) {
	routes := []string{
		"/foo",
		"/foo/bar",
		"/ping/pong/",
	}
	root := newTestTree(routes)
	cases := []testCase{
		{
			path: "/",
			_404: true,
		},
		{
			path: "/foo/bar/baz",
			_404: true,
		},
		{
			path: "/ping/pong",
			_404: true,
		},
		{
			path:    "/foo",
			handler: 0,
		},
		{
			path:    "/foo/bar",
			handler: 1,
		},
		{
			path:    "/ping/pong/xyz",
			handler: 2,
		},
	}
	dotest(t, root, cases)
}

func TestLookup_Dynamic(t *testing.T) {
	routes := []string{
		"/:foo",
		"/:foo/:bar",
		"/ping/pong",
	}
	root := newTestTree(routes)
	cases := []testCase{
		{
			path:    "/",
			handler: 0,
			vars:    Vars{{"foo", ""}},
		},
		{
			path:    "/bob",
			handler: 0,
			vars:    Vars{{"foo", "bob"}},
		},
		{
			path:    "/bob/alice",
			handler: 1,
			vars:    Vars{{"foo", "bob"}, {"bar", "alice"}},
		},
	}
	dotest(t, root, cases)
}

func dotest(t *testing.T, root *node, cases []testCase) {
	for _, v := range cases {
		chunks := bytes.Split([]byte(v.path), []byte{'/'})
		tar, vars := root.lookup(chunks, testAlloc)
		if v._404 {
			if tar != nil && tar.handler != nil {
				t.Fatalf("[%s] bad node, want <nil> got %s", v.path, tar)
			}
		} else {
			if tar == nil {
				t.Fatalf("[%s] nil node", v.path)
			}
			if tar.handler != v.handler {
				t.Fatalf("[%s] bad handler. want %d got %v", v.path, v.handler, tar.handler)
			}
			if !reflect.DeepEqual(vars, v.vars) {
				t.Fatalf("[%s] bad variables. want %v got %v", v.path, v.vars, vars)
			}
		}
	}
}

func TestLookup_Priority(t *testing.T) {
	root := new(node)

	h0 := testHandler(0)
	h1 := testHandler(1)
	h2 := testHandler(2)

	path := testParse("/bob/alice")

	// wildcard
	root.insert(testParse("/"), 0, h2)
	node, _ := root.lookup(path, testAlloc)
	if node.handler != h2 {
		t.Fatalf("want %v got %v", h2, node.handler)
	}

	// dynamic
	root.insert(testParse("/:foo/:bar"), 0, h1)
	node, _ = root.lookup(path, testAlloc)
	if node.handler != h1 {
		t.Fatalf("want %v got %v", h1, node.handler)
	}

	// static
	root.insert(testParse("/bob/alice"), 0, h0)
	node, _ = root.lookup(path, testAlloc)
	if node.handler != h0 {
		t.Fatalf("want %v got %v", h0, node.handler)
	}
}

func testParse(s string) [][]byte {
	return bytes.Split([]byte(s), []byte{'/'})
}
