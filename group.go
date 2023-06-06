package hr

import (
	"net/http"
	"path"
	"sync"
	"unsafe"
)

type Group struct {
	prev    *Group
	prefix  string
	trees   [12]*node
	chunks  *sync.Pool
	plugins []Plugin
}

func (g *Group) Prefix(prefix string, plugins ...Plugin) *Group {
	return &Group{
		prev:    g,
		prefix:  prefix,
		plugins: plugins,
	}
}

func (g *Group) Handle(method, route string, handler Handler, plugins ...Plugin) {
	if g.prev == nil {
		g.handle(method, route, handler, plugins...)
		return
	}
	ps := plugins
	if len(plugins) > 0 {
		ps = make([]Plugin, 0, len(g.plugins)+len(plugins))
		ps = append(ps, g.plugins...)
		ps = append(ps, plugins...)
	}
	g.prev.Handle(method, path.Join(g.prefix, route), handler, ps...)
}

func (g *Group) handle(method, route string, h Handler, ps ...Plugin) {
	i := hash(method)
	if g.trees[i] == nil {
		g.trees[i] = &node{}
	}
	if len(route) == 0 {
		route = "/"
	}
	// compose handler plugins.
	for j := len(ps) - 1; j >= 0; j-- {
		h = ps[j](h)
	}
	// compose root handler plugins.
	for j := len(g.plugins) - 1; j >= 0; j-- {
		h = g.plugins[j](h)
	}

	chunks := g.chunks.Get().([][]byte)
	unsafeParse(path.Join(g.prefix, route), &chunks)
	g.trees[i].insert(chunks, 0, h)
}

func (g *Group) GET(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodGet, route, h, ps...)
}

func (g *Group) POST(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodPost, route, h, ps...)
}

func (g *Group) PUT(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodPut, route, h, ps...)
}

func (g *Group) PATCH(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodPatch, route, h, ps...)
}

func (g *Group) DELETE(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodDelete, route, h, ps...)
}

func (g *Group) HEAD(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodHead, route, h, ps...)
}

func (g *Group) OPTIONS(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodOptions, route, h, ps...)
}

func (g *Group) CONNECT(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodConnect, route, h, ps...)
}

func (g *Group) TRACE(route string, h HandlerFunc, ps ...Plugin) {
	g.Handle(http.MethodTrace, route, h, ps...)
}

// G|ET|
// P|OS|T
// P|UT|
// P|AT|CH
// H|EA|D
// O|PT|IONS
// C|ON|NECT
// T|RA|CE
// D|EL|ETE
func hash(s string) int {
	return int(s[1]+s[2]) % 18
}

func unsafeParse(s string, c *[][]byte) {
	b := unsafeAtobs(s)

	i, j := 0, 0
	for ; j < len(b); j++ {
		if b[j] == '/' {
			*c = append(*c, b[i:j])
			i = j + 1
		}
	}
	*c = append(*c, b[i:])
}

func unsafeAtobs(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			int
		}{s, len(s)},
	))
}
