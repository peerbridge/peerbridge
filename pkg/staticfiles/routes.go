package staticfiles

import (
	"mime"
	"net/http"
	"path/filepath"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

var fileServer = http.FileServer(http.Dir("./static"))

func contentType(file string) string {
	// We use a built in table of the common types since
	// the system TypeByExtension might be unreliable.
	// But if we don't know, we delegate to the system.

	ext := filepath.Ext(file)
	switch ext {
	case ".htm", ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	default:
		return mime.TypeByExtension(ext)
	}
}

// Serve static files.
func serve(w http.ResponseWriter, r *http.Request) {
	t := contentType(r.URL.Path)
	w.Header().Set("Content-Type", t)
	fileServer.ServeHTTP(w, r)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/", serve)
	return
}
