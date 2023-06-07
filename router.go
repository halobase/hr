package hr

import (
	"net/http"
	"sync"
)

type Handler interface {
	ServeHTTP(*Ctx) error
}

type HandlerFunc func(*Ctx) error

func (f HandlerFunc) ServeHTTP(c *Ctx) error { return f(c) }

func NopHandlerFunc(c *Ctx) error { return nil }

type Var struct {
	Key   string
	Value string
}

type Vars []Var

func (vs Vars) Get(key string) string {
	for _, v := range vs {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

type Plugin func(Handler) Handler

type Router struct {
	Group
	chunks  sync.Pool
	context sync.Pool
	vars    sync.Pool
}

func Default(plugins ...Plugin) *Router {
	return New("", plugins...)
}

func New(prefix string, plugins ...Plugin) *Router {
	router := &Router{
		vars:    sync.Pool{New: func() any { return make(Vars, 0, 32) }},
		chunks:  sync.Pool{New: func() any { return make([][]byte, 0) }},
		context: sync.Pool{New: func() any { return new(Ctx) }},
	}
	group := Group{
		prefix:  prefix,
		plugins: plugins,
		chunks:  &router.chunks,
	}
	router.Group = group

	// we have to register an internal handler for method options
	// to make plugins using such method work right :(
	router.handle(http.MethodOptions, "/", HandlerFunc(NopHandlerFunc))
	return router
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "*" {
		if req.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	path := req.URL.Path
	if len(path) == 0 {
		path = "/"
	}

	i := hash(req.Method)
	tree := r.trees[i]

	if tree == nil {
		http.NotFound(w, req)
		return
	}

	alloc := func() Vars { return r.vars.Get().(Vars) }
	chunks := r.chunks.Get().([][]byte)
	unsafeParse(path, &chunks)
	node, vars := tree.lookup(chunks, alloc)

	if node != nil && node.handler != nil {
		h := node.handler
		ctx := r.context.Get().(*Ctx)
		ctx.Context = req.Context()
		ctx.req = req
		ctx.rw = w
		ctx.vars = vars

		if err := h.ServeHTTP(ctx); err != nil {
			// TODO: we should log the error if failed to send the response
			ParseError(err).WriteTo(w)
		}
		// put ctx back to the context pool.
		r.context.Put(ctx)
	} else {
		http.NotFound(w, req)
	}

	// put vars and chunks back to their pools.
	chunks = chunks[:0] // reset
	r.chunks.Put(chunks)
	if vars == nil {
		return
	}
	vars = vars[:0]
	r.vars.Put(vars)
}
