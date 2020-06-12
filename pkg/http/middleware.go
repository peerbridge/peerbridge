package http

import "net/http"

// `Middleware` is a function which receives an http.Handler and returns another http.Handler.
// This can be done using a closure where the middle executes operations on the http.ResponseWriter
// and http.Request and calls the calls the handler passed as parameter
type Middleware func(http.Handler) http.Handler
