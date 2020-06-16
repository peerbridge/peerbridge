package http

import (
	"net/http"
)

// Route stores information to match a request and build URLs.
type Route struct {
	Method  string
	Pattern string
	Handler http.Handler
}
