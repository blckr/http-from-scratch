package mux

import (
	"context"
	"net/http"
	"sort"
	"strings"
)

type SimpleMux struct {
	root *node
}

func NewSimpleMux() *SimpleMux {
	return &SimpleMux{
		root: newNode(""),
	}
}

func (m *SimpleMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	method := ""
	path := pattern

	//Check if there is a method prefix
	parts := strings.SplitN(strings.TrimSpace(pattern), " ", 2)
	if len(parts) == 2 {
		method = strings.ToUpper(parts[0])
		path = parts[1]
	}

	segments := splitPath(path)
	m.root.insert(method, segments, handler)
}

func (m *SimpleMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	segments := splitPath(r.URL.Path)
	params := make(map[string]string)

	matchedNode := m.root.search(segments, params)

	if matchedNode == nil || len(matchedNode.handlers) == 0 {
		http.NotFound(w, r)
		return
	}

	ctx := context.WithValue(r.Context(), paramsKey, params)
	reqWithContext := r.WithContext(ctx)

	m.dispatchMethod(w, reqWithContext, matchedNode.handlers)
}

// Checks inside a Path if the Method exists
func (m *SimpleMux) dispatchMethod(w http.ResponseWriter, r *http.Request, methods map[string]http.HandlerFunc) {

	// Method found
	if handler, ok := methods[r.Method]; ok {
		handler(w, r)
		return
	}

	// Catch ALL
	if handler, ok := methods[""]; ok {
		handler(w, r)
		return
	}

	// Write not allowed back in header
	allowed := make([]string, 0, len(methods))
	for method := range methods {
		if method != "" {
			allowed = append(allowed, method)
		}
	}
	sort.Strings(allowed)

	w.Header().Set("Allow", strings.Join(allowed, ", "))
	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
}

// redirectURL
func redirectURL(w http.ResponseWriter, r *http.Request, targetPath string) {
	target := targetPath
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
