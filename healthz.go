package main

import (
	"net/http"
)

var healthz http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	// set headers
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		panic("")
	}
}
