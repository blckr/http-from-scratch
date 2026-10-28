package mux

import (
	"net/http"
	"sort"
	"strings"
)

type SimpleMux struct {
	routes map[string]map[string]http.HandlerFunc
}

func NewSimpleMux() *SimpleMux {
	return &SimpleMux{
		routes: make(map[string]map[string]http.HandlerFunc),
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

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if m.routes[path] == nil {
		m.routes[path] = make(map[string]http.HandlerFunc)
	}

	// Empty means catch all
	m.routes[path][method] = handler
}

func (m *SimpleMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if methods, found := m.routes[path]; found {
		m.dispatchMethod(w, r, methods)
		return
	}

	if !strings.HasSuffix(path, "/") {
		if _, found := m.routes[path+"/"]; found {
			redirectURL(w, r, path+"/")
			return
		}
	}

	if strings.HasSuffix(path, "/") && path != "/" {
		pathWithoutSlash := strings.TrimSuffix(path, "/")
		if _, found := m.routes[pathWithoutSlash]; found {
			redirectURL(w, r, pathWithoutSlash)
			return
		}
	}

	http.NotFound(w, r)
}

// Checks inside a Path if the Method exists
func (m *SimpleMux) dispatchMethod(w http.ResponseWriter, r *http.Request, methods map[string]http.HandlerFunc) {

	if handler, ok := methods[r.Method]; ok {
		handler(w, r)
		return
	}

	if handler, ok := methods[""]; ok {
		handler(w, r)
		return
	}

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

func redirectURL(w http.ResponseWriter, r *http.Request, targetPath string) {
	target := targetPath
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
