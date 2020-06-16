package http

import (
	"net/http"
)

type Router struct {
	mux         *http.ServeMux
	routes      []Route
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{http.NewServeMux(), make([]Route, 0), make([]Middleware, 0)}
}

// `Use` appends a Middleware to the chain.
// Middleware can be used to intercept or otherwise modify requests and/or responses,
// and are executed in the order that they are applied to the Router.
func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// `Add` adds new Routes to the Router.
// Routes handle requests and are executed according
// to their pattern and supported methods.
func (r *Router) Add(routes ...Route) {
	r.routes = append(r.routes, routes...)
	for _, route := range routes {
		r.mux.Handle(route.Pattern, route.Handler)
	}
}

// `ServeHTTP` handles the http request and writes the response
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Set the initial handler function to serve the request
	// this will simply be calling ServeHTTP on the multiplexer
	// ServeMux where all routes where registered during `Add`
	var handler http.Handler
	handler = router.mux

	// apply middlewares
	for _, middleware := range router.middlewares {
		// middleware is a function in the form (http.Handler) -> http.Handler
		// thus returning an instance of the http.Handler interface
		handler = middleware(handler)
	}

	// write the response calling ServeHTTP on the http.Handler interface
	handler.ServeHTTP(w, r)
}
