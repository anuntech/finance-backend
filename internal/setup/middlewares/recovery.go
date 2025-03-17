package middlewares

import (
	"log"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ERRO GRAVE: %v\n", err)
				log.Printf("Stack trace: %s\n", debug.Stack())

				if w.Header().Get("Content-Type") == "" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"Erro interno no servidor"}`))
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
