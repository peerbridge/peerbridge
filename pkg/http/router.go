package http

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

const DefaultPort = "8080"

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

// `Get` adds a new Route with http method "GET" to the Router.
func (r *Router) Get(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.Add(Route{Method: http.MethodGet, Pattern: pattern, Handler: http.HandlerFunc(handlerFunc)})
}

// `Post` adds a new Route with http method "POST" to the Router.
func (r *Router) Post(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.Add(Route{Method: http.MethodPost, Pattern: pattern, Handler: http.HandlerFunc(handlerFunc)})
}

// `Put` adds a new Route with http method "PUT" to the Router.
func (r *Router) Put(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.Add(Route{Method: http.MethodPut, Pattern: pattern, Handler: http.HandlerFunc(handlerFunc)})
}

// `Patch` adds a new Route with http method "PATCH" to the Router.
func (r *Router) Patch(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.Add(Route{Method: http.MethodPatch, Pattern: pattern, Handler: http.HandlerFunc(handlerFunc)})
}

// `Mount` adds a new Router which handles requests on the specified pattern.
func (r *Router) Mount(pattern string, subRouter *Router) {
	r.mux.Handle(pattern+"/", http.StripPrefix(pattern, subRouter))
}

// `ServeHTTP` handles the http request and writes the response
func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func GetServerPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return DefaultPort
}

// Serve accepts incoming HTTP connections on the listener l
func (r *Router) Serve(l net.Listener) error {
	return http.Serve(l, r)
}

// ServeTLS accepts incoming HTTPS connections on the listener l
func (r *Router) ServeTLS(l net.Listener, certFile, keyFile string) error {
	return http.ServeTLS(l, r, certFile, keyFile)
}

// `ListenAndServe` launches the default http server
func (r *Router) ListenAndServe() error {
	// generate server address
	addr := fmt.Sprintf(":%s", GetServerPort())
	srv := http.Server{Addr: addr, Handler: r}
	return srv.ListenAndServe()
}
