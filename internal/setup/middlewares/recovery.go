package middlewares

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC ERROR: %v\n", err)
				stack := debug.Stack()
				log.Printf("Stack trace: %s\n", stack)

				w.WriteHeader(http.StatusInternalServerError)
				_, err := io.Copy(w, bytes.NewReader(stack))
				if err != nil {
					return
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
