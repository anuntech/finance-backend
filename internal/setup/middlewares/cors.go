package middlewares

import (
	"net/http"
	"os"
	"strings"
)

func CorsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := make(map[string]bool)

	// Lê as origens permitidas da variável de ambiente
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv != "" {
		origins := strings.Split(allowedOriginsEnv, ",")
		for _, origin := range origins {
			allowedOrigins[strings.TrimSpace(origin)] = true
		}
	} else {
		// Fallback para valores padrão se a variável de ambiente não estiver definida
		allowedOrigins = map[string]bool{
			"https://anun.tech":                       true,
			"http://localhost:3000":                   true,
			"http://localhost:3001":                   true,
			"https://anuntech.online":                 true,
			"https://finance-company.anuntech.online": true,
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, workspaceid, X-Requested-With, Accept")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
