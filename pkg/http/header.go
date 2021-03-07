package http

import (
	"net/http"
)

// Configure the header for http responses.
func Header(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Add Content-Type header, the Api will always return json responses
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		// Add X-XSS-Protection header
		w.Header().Add("X-XSS-Protection", "1; mode=blockFilter")

		// Add X-Content-Type-Options header
		w.Header().Add("X-Content-Type-Options", "nosniff")

		// Prevent page from being displayed in an iframe
		w.Header().Add("X-Frame-Options", "DENY")

		inner.ServeHTTP(w, r)

	})
}
