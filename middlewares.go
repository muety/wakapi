package main

import "net/http"

func AuthenticateMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(w, r)
}
