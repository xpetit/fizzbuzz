package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xpetit/fizzbuzz/handlers"
)

func main() {
	// Setup signal handler
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Use microseconds resolution for logs
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	// Configure HTTP server with sane defaults
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	api := http.NewServeMux()
	srv := http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handlers.Logger(log.Default())(api),
	}

	// Configure HTTP routing
	var s handlers.Stats
	api.HandleFunc("/api/v1/fizzbuzz", s.HandleFizzBuzz)
	api.HandleFunc("/api/v1/statistics", s.HandleStats)

	// Spawn a goroutine that waits for a termination signal and then stops the HTTP server
	go func() {
		<-sig
		log.Println("shutting down HTTP server")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()

	// Start the HTTP server
	log.Println("listening on port:", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalln(err)
	}
}
