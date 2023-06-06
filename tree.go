package hr

import (
	"bytes"
	"fmt"
	"sort"
)

type nodeKind uint8

const (
	nodeUnknown nodeKind = iota
	nodeStatic
	nodeDynamic
	nodeWildcard

	vardec byte = ':'
)

type nodes []*node

func (a nodes) Len() int           { return len(a) }
func (a nodes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodes) Less(i, j int) bool { return a[i].kind < a[j].kind }

type node struct {
	kind     nodeKind
	chunk    []byte
	children nodes
	wildcard *node
	handler  Handler
}

func (n node) String() string {
	return fmt.Sprintf("node(kind=%d chunk=%s children=%v)", n.kind, string(n.chunk), n.children)
}

func (n *node) insert(chunks [][]byte, height int, handler Handler) {
	if height == len(chunks) {
		if n.handler != nil {
			panic("redefining route")
		}
		n.handler = handler
		return
	}

	if n.children == nil {
		n.children = make(nodes, 0)
	}

	chunk := chunks[height]
	child := n.child(chunk)

	if child == nil {
		child = &node{
			kind:  chunkToKind(chunk),
			chunk: chunk,
		}

		n.children = append(n.children, child)
		sort.Sort(n.children)

		if child.kind == nodeWildcard {
			n.wildcard = child
		}
	}

	child.insert(chunks, height+1, handler)
}

func (n node) child(chunk []byte) *node {
	for _, child := range n.children {
		if bytes.Equal(child.chunk, chunk) {
			return child
		}
	}
	return nil
}

func chunkToKind(chunk []byte) nodeKind {
	switch {
	case len(chunk) == 0:
		return nodeWildcard
	case chunk[0] == vardec:
		return nodeDynamic
	default:
		return nodeStatic
	}
}

func (n node) lookup(chunks [][]byte, alloc func() Vars) (*node, Vars) {
	var wild *node // nearest wildcard node along the path.
	var vars Vars  // route variables to be parsed.

	// let's start from the cur node.
	cur := &n
	found := false

	for _, chunk := range chunks {
		// clear the found flag
		found = false
		// remember the nearest non-nil wildcard node
		if cur.wildcard != nil {
			wild = cur.wildcard
		}

		for _, child := range cur.children {
			switch {
			case child.kind == nodeStatic && bytes.Equal(chunk, child.chunk):
			case child.kind == nodeDynamic:
				// do not forget to allocate memory for the route variables.
				if vars == nil {
					vars = alloc()
				}
				// expand the variable slice within preallocated capacity
				// like how julienschmidt/httprouter does :)
				i := len(vars)
				vars = vars[:i+1]
				vars[i] = Var{
					Key:   string(child.chunk[1:]),
					Value: string(chunk),
				}
			case child.kind == nodeWildcard:
			default:
				// try the next child if the current child did not
				// match the chunk.
				continue
			}
			// we found one!
			cur = child
			found = true
			// now we can jump to the next chunk to see if it matches
			// the child node we just found.
			goto next
		}
		// unfortunately not one single node matches the current chunk
		// we have to stop from here and fallback to the nearest wildcard
		// node.
		if !found {
			cur = wild
			return cur, vars
		}
	next:
	}
	// but if the one we just found has no handler which mean
	// we found it before reaching a leaf node, we still have
	// to fallback to the nearest wildcard node.
	if cur.handler == nil {
		cur = wild
	}
	return cur, vars
}
