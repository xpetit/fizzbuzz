package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/xpetit/fizzbuzz/v5/handlers"
	"github.com/xpetit/fizzbuzz/v5/stats"
)

func Main() error {
	// Setup signal handler
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Use microseconds resolution for logs
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	// Parse config flags
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	defaultDBFile := filepath.Join(dir, "fizzbuzz", "data.db")
	dbFile := flag.String("db", defaultDBFile, `The path to the SQLite database file. Special values:
    off         to disable SQLite (stats are kept in memory)
    :memory:    to get an in-memory SQLite database
`)
	flag.Parse()

	// Initialize stats service
	var statsService stats.Service
	if *dbFile == "off" {
		statsService = &stats.Memory{}
	} else {
		if !strings.Contains(*dbFile, ":memory:") {
			if err := os.MkdirAll(filepath.Dir(*dbFile), 0700); err != nil {
				return err
			}
		}
		statsDB, err := stats.Open(ctx, *dbFile)
		if err != nil {
			return err
		}
		defer statsDB.Close()
		statsService = statsDB
	}

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
	fb := handlers.Fizzbuzz{Stats: statsService}
	api.HandleFunc("/api/v2/fizzbuzz", fb.Handle)
	api.HandleFunc("/api/v2/fizzbuzz/stats", fb.HandleStats)

	// Spawn a goroutine that waits for a termination signal and then stops the HTTP server
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server")
		shutdownErr <- srv.Shutdown(context.Background())
	}()

	// Start the HTTP server
	log.Println("Listening on port:", port, "You can customize it with PORT environment variable")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on port %s: %w", port, err)
	}

	// Wait for the HTTP server to shutdown
	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("could not shutdown HTTP server: %w", err)
	}

	if statsService, ok := statsService.(*stats.DB); ok {
		return statsService.Close()
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}
