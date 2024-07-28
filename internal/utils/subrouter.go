package utils

import (
	"net/http"
	"strings"
)

func Subrouter(router *http.ServeMux, route string) *http.ServeMux {
	sr := http.NewServeMux()
	route = strings.TrimSuffix(route, "/")
	router.Handle(route, removePrefix(sr, route))
	router.Handle(route+"/", removePrefix(sr, route))
	return sr
}

func removePrefix(h http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		r.URL.Path = "/" + strings.TrimPrefix(strings.TrimPrefix(path, prefix), "/")
		h.ServeHTTP(w, r)
		r.URL.Path = path
	})
}
