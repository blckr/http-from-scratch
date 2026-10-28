package mux

import (
	"net/http"
	"strings"
)

type node struct {
	part      string
	children  []*node
	wildcard  bool
	paramName string
	handlers  map[string]http.HandlerFunc
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

func newNode(part string) *node {
	wildcard := strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}")
	paramName := ""

	if wildcard {
		paramName = strings.Trim(part, "{}")
	}

	return &node{
		part:      part,
		children:  make([]*node, 0),
		wildcard:  wildcard,
		paramName: paramName,
		handlers:  make(map[string]http.HandlerFunc),
	}
}

func (n *node) insert(method string, segments []string, handler http.HandlerFunc) {

	if len(segments) == 0 {
		n.handlers[method] = handler
		return
	}

	segment := segments[0]
	var child *node

	for _, c := range n.children {
		if c.part == segment {
			child = c
			break
		}
	}

	if child == nil {
		child = newNode(segment)
		n.children = append(n.children, child)
	}

	child.insert(method, segments[1:], handler)
}

func (n *node) search(segments []string, params map[string]string) *node {
	if len(segments) == 0 {
		return n
	}

	segment := segments[0]
	var wildChild *node

	for _, child := range n.children {
		if child.wildcard {
			wildChild = child
		} else if child.part == segment {
			if res := child.search(segments[1:], params); res != nil {
				return res
			}
		}
	}

	if wildChild != nil {
		params[wildChild.paramName] = segment
		if res := wildChild.search(segments[1:], params); res != nil {
			return res
		}
		delete(params, wildChild.paramName)
	}
	return nil
}

type contextKey string

const paramsKey contextKey = "params"

// PathValue extracts a dynamic URL path parameter (e.g. "{id}") from the request context.
// It returns the value associated with the specified key, or an empty string if the
// key does not exist or if no parameters were set for this request.
func PathValue(r *http.Request, key string) string {
	params, ok := r.Context().Value(paramsKey).(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}
