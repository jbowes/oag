package translator

// trie organizes Operations by path prefix, so they can be grouped into appropriate
// client handlers.
//
// Paths are split on '/', and marked as a literal or param value. We can use the
// path ordering of params to construct method argument order.

import (
	"sort"
	"unicode/utf8"

	"github.com/jbowes/oag/openapi/v2"
)

type (
	token interface {
		value() string
	}
	literal string
	param   string
)

func (l literal) value() string { return string(l) }
func (p param) value() string   { return "{" + string(p) + "}" }

type node struct {
	prefix   token
	handlers map[string]*v2.Operation

	// The two child types
	literals []*node
	params   []*node
}

func (n *node) add(path string, pi *v2.PathItem) {
	tokens := tokenize(path)
	tok := tokens[0]
	switch tok.(type) {
	case param:
		n.params = addChild(path, tokens, n.params, pi)
	default:
		n.literals = addChild(path, tokens, n.literals, pi)
	}
}

func addChild(path string, tokens []token, children []*node, pi *v2.PathItem) []*node {
	tok := tokens[0]
	for _, child := range children {
		if child.prefix != tok {
			continue
		}

		if len(tokens) == 1 {
			child.setHandlers(pi)
		} else {
			child.add(path[len(tok.value()):], pi)
		}
		return children
	}

	child := &node{prefix: tok}
	if len(tokens) > 1 {
		child.add(path[len(tok.value()):], pi)
	} else {
		child.setHandlers(pi)
	}
	children = append(children, child)

	sort.Slice(children, func(i, j int) bool {
		return children[i].prefix.value() < children[j].prefix.value()
	})
	return children
}

func (n *node) setHandlers(pi *v2.PathItem) {
	n.handlers = make(map[string]*v2.Operation)
	if pi.Get != nil {
		n.handlers["Get"] = pi.Get
	}
	if pi.Post != nil {
		n.handlers["Post"] = pi.Post
	}
	if pi.Put != nil {
		n.handlers["Put"] = pi.Put
	}
	if pi.Patch != nil {
		n.handlers["Patch"] = pi.Patch
	}
	if pi.Delete != nil {
		n.handlers["Delete"] = pi.Delete
	}
	if pi.Options != nil {
		n.handlers["Options"] = pi.Options
	}
	if pi.Head != nil {
		n.handlers["Head"] = pi.Head
	}
}

type visited struct {
	path []token
	n    *node
}

// visit returns a channel that iterates over all nodes in depth first preorder
// traversal.
func (n *node) visit() <-chan *visited {
	c := make(chan *visited)

	var cur *visited
	stack := []*visited{{[]token{n.prefix}, n}}
	go func() {
		for len(stack) > 0 {
			end := len(stack) - 1
			cur, stack = stack[end], stack[:end]

			c <- cur
			for i := 0; i < len(cur.n.literals); i++ {
				child := cur.n.literals[i]
				stack = append(stack, &visited{append(cur.path, child.prefix), child})
			}

			for i := 0; i < len(cur.n.params); i++ {
				child := cur.n.params[i]
				stack = append(stack, &visited{append(cur.path, child.prefix), child})
			}
		}

		close(c)
	}()

	return c
}

// XXX handle illegal values by panicing
func tokenize(path string) []token {
	var toks []token

	rc := runeChan(path)

	start, ok := <-rc

	for ok {
		var tok token
		switch start {
		case '{':
			tok, start, ok = tokenizeParam(start, rc)
		default:
			tok, start, ok = tokenizeLiteral(start, rc)
		}
		toks = append(toks, tok)
	}

	return toks
}

func tokenizeParam(start rune, rc <-chan rune) (token, rune, bool) {
	val := ""
	for {
		c, ok := <-rc
		switch {
		case !ok, c == '}':
			c, ok = <-rc
			return param(val), c, ok
		default:
			val += string(c)
		}
	}
}

func tokenizeLiteral(start rune, rc <-chan rune) (token, rune, bool) {
	val := string(start)
	if start == '/' {
		c, ok := <-rc
		return literal(val), c, ok
	}

	for {
		c, ok := <-rc
		switch {
		case !ok, c == '{', c == '/':
			return literal(val), c, ok
		default:
			val += string(c)
		}
	}
}

// XXX returning something that can be closed would let this be set up as
// something without a buffer size, so the chan can close.
func runeChan(val string) <-chan rune {
	rc := make(chan rune, utf8.RuneCountInString(val))
	go func() {
		for _, r := range val {
			rc <- r
		}
		close(rc)
	}()

	return rc
}
