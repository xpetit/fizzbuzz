package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xpetit/fizzbuzz/v2/handlers"
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
	api.HandleFunc("/api/v2/fizzbuzz", s.HandleFizzBuzz)
	api.HandleFunc("/api/v2/fizzbuzz/stats", s.HandleStats)
	// v1 compatibility
	api.HandleFunc("/api/v1/fizzbuzz", s.HandleFizzBuzz)
	api.HandleFunc("/api/v1/fizzbuzz/stats", s.HandleStatsV1)

	// Spawn a goroutine that waits for a termination signal and then stops the HTTP server
	go func() {
		<-sig
		log.Println("Shutting down HTTP server")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()

	// Start the HTTP server
	log.Println("Listening on port:", port, "You can customize it with PORT environment variable")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalln(err)
	}
}
