package http

import (
	"net/http"
)

// Route stores information to match a request and build URLs.
type Route struct {
	method  string
	pattern string
	handler http.Handler
}
