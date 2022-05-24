package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xpetit/x"

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
	var (
		dbFile string
		host   string
		port   uint
	)
	{
		dir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		defaultDBFile := filepath.Join(dir, "fizzbuzz", "data.db")
		flag.StringVar(&dbFile, "db", defaultDBFile, x.MultiLines(`
			The path to the SQLite database file. Special values:
				off         to disable SQLite (stats are kept in memory)
				:memory:    to get an in-memory SQLite database

		`))
		flag.StringVar(&host, "host", "127.0.0.1", "address to bind to")
		flag.UintVar(&port, "port", 8080, "listening port")
		flag.Parse()
	}

	// Initialize stats service
	var statsService stats.Service
	if dbFile == "off" {
		statsService = stats.Memory()
	} else {
		if !strings.Contains(dbFile, ":memory:") {
			if err := os.MkdirAll(filepath.Dir(dbFile), 0o700); err != nil {
				return err
			}
		}
		db, err := stats.OpenDB(ctx, dbFile)
		if err != nil {
			return err
		}
		defer db.Close()
		statsService = db
	}

	// Configure HTTP server with sane defaults
	api := http.NewServeMux()
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	srv := http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handlers.Logger(log.Default())(api),
	}

	// Configure HTTP routing
	fb := handlers.Fizzbuzz(statsService)
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
	log.Println("Listening on", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen on port %d: %w", port, err)
	}

	// Wait for the HTTP server to shutdown
	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if c, ok := statsService.(io.Closer); ok {
		return c.Close()
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}
