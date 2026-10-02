package middleware

import (
	"net/http"
	"strings"

	"changeme/backend"
)

func PosterMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/posters") {
			if backend.PosterDir == "" {
				http.NotFound(w, r)
			}

			// serve file
			http.StripPrefix("/posters", http.FileServer(http.Dir(backend.PosterDir))).ServeHTTP(w, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}
