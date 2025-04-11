package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ERRO CRÍTICO: %v\n", err)
				stack := debug.Stack()
				log.Printf("Rastreamento da pilha: %s\n", stack)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := map[string]string{
					"error": "Desculpe, encontramos um problema inesperado. Nossa equipe técnica foi notificada e já está trabalhando para resolver. Por favor, tente novamente em alguns instantes.",
				}

				json.NewEncoder(w).Encode(errorResponse)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
