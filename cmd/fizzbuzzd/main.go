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

	"github.com/xpetit/fizzbuzz/v5/handlers"
	"github.com/xpetit/fizzbuzz/v5/stats"
)

type Config struct {
	DBFile  string
	Addr    string
	logging bool
}

func (c *Config) Run(ctx context.Context) error {
	// Initialize stats service
	var statsService stats.Service
	if c.DBFile == "off" {
		statsService = stats.Memory()
	} else {
		if !strings.Contains(c.DBFile, ":memory:") {
			if err := os.MkdirAll(filepath.Dir(c.DBFile), 0o700); err != nil {
				return err
			}
		}
		log.Println("Using database file:", c.DBFile)
		db, err := stats.OpenDB(ctx, c.DBFile)
		if err != nil {
			return err
		}
		defer db.Close()
		statsService = db
	}

	// Configure HTTP server
	api := http.NewServeMux()
	fb := handlers.Fizzbuzz(statsService)
	api.HandleFunc("/api/v2/fizzbuzz", fb.Handle)
	api.HandleFunc("/api/v2/fizzbuzz/stats", fb.HandleStats)
	api.HandleFunc("/api/v2/ready", func(http.ResponseWriter, *http.Request) {})
	srv := http.Server{
		Addr:         c.Addr,
		Handler:      api,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if c.logging {
		srv.Handler = handlers.Logger(log.Default())(srv.Handler)
	}

	// Spawn a goroutine that waits for a termination signal and then stops the HTTP server
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server")
		shutdownErr <- srv.Shutdown(context.Background())
	}()

	// Start the HTTP server
	log.Println("Listening on", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listening on %s: %w", srv.Addr, err)
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

func run() error {
	// Setup signal handler
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Use microseconds resolution for logs
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	// Parse config flags
	var c Config
	var host string
	var port int
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	defaultDBFile := filepath.Join(dir, "fizzbuzz", "data.db")
	flag.BoolVar(&c.logging, "logging", true, "Enable HTTP logging")
	flag.StringVar(&c.DBFile, "db", defaultDBFile, `The path to the SQLite database file. Special values:
	off         to disable SQLite (stats are kept in memory)
	:memory:    to get an in-memory SQLite database
`)
	flag.StringVar(&host, "host", "127.0.0.1", "address to bind to")
	flag.IntVar(&port, "port", 8080, "listening port")
	flag.Parse()

	c.Addr = net.JoinHostPort(host, strconv.Itoa(port))

	return c.Run(ctx)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
