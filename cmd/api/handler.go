package api

import (
	"net/http"
)

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type Handler struct {
	H func(w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(w, r)

	if err != nil {
		switch e := err.(type) {
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, e.Error(),
				http.StatusInternalServerError)
		}
	}
}
