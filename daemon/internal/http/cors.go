package http

import (
	"net/http"
	"strings"
)

func corsMiddleware(
	allowedOrigins []string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin == "" {
				next.ServeHTTP(w, r)

				return
			}

			if !isAllowedOrigin(
				origin,
				allowedOrigins,
			) {
				next.ServeHTTP(w, r)

				return
			}

			w.Header().Set(
				"Access-Control-Allow-Origin",
				origin,
			)

			w.Header().Set(
				"Vary",
				"Origin",
			)

			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, OPTIONS",
			)

			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type",
			)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

func isAllowedOrigin(
	origin string,
	allowedOrigins []string,
) bool {
	host := stripScheme(origin)

	for _, allowedOrigin := range allowedOrigins {
		if host == stripScheme(allowedOrigin) {
			return true
		}
	}

	return false
}

func stripScheme(origin string) string {
	if _, rest, found := strings.Cut(origin, "://"); found {
		return rest
	}

	return origin
}
