package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	srv := http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	http.HandleFunc("/v1/fizzbuzz", func(rw http.ResponseWriter, r *http.Request) {})

	// Spawn a goroutine that waits for a termination signal and then gracefully stops the HTTP server
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
