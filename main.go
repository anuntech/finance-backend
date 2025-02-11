package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anuntech/finance-backend/internal/setup"
	"github.com/anuntech/finance-backend/internal/setup/config"
)

func corsMiddleware(next http.Handler) http.Handler {
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
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	config.LoadEnvFile(".env")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("server is running with port", port)

	// Adiciona o middleware CORS ao servidor
	handler := corsMiddleware(setup.Server())

	sm := http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		err := sm.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Println("received terminate, graceful shutdown", sig)

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sm.Shutdown(tc)
}
