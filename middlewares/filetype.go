package middlewares

import (
	"net/http"
	"strings"
)

type SuffixFilterMiddleware struct {
	handler     http.Handler
	filterTypes []string
}

func NewFileTypeFilterMiddleware(filter []string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &SuffixFilterMiddleware{
			handler:     h,
			filterTypes: filter,
		}
	}
}

func (f *SuffixFilterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.ToLower(r.URL.Path)
	for _, t := range f.filterTypes {
		if strings.HasSuffix(path, strings.ToLower(t)) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("403 forbidden"))
			return
		}
	}
	f.handler.ServeHTTP(w, r)
}
