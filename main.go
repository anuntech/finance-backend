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
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
)

func main() {
	config.LoadEnvFile(".env")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("server is running with port", port)

	handler := middlewares.CorsMiddleware(setup.Server())

	sm := http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
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
