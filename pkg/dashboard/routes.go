package dashboard

import (
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

func dashboardView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./templates/dashboard.html")
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/", dashboardView)
	router.Get("/ws", BindNewClient)
	return
}
