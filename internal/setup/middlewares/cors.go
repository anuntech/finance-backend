package middlewares

import "net/http"

func CorsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"https://anun.tech":                 true,
		"http://localhost:3000":             true,
		"http://localhost:3001":             true,
		"https://anuntech.com":              true,
		"https://finance-company.anun.tech": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, workspaceid, X-Requested-With, Accept")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
