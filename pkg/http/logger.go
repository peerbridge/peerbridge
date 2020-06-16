package http

import (
	"log"
	"net/http"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
)

func Logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s",
			color.Sprintf(r.Method, color.InfoColor),
			color.Sprintf(r.RequestURI, color.NoticeColor),
			color.Sprintf(time.Since(start).String(), color.DebugColor),
		)
	})
}
