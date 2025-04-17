package middlewares

import "net/http"

func AllowCacheHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=60")
		next.ServeHTTP(w, r)
	})
}
