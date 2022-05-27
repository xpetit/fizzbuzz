package handlers

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// Logger is a basic HTTP middleware that logs queries with the following format:
// {remote IP address} {time taken with microsecond resolution} {HTTP method} {URL}
// Example:
// 	// 172.17.0.1 0.000003s GET /api/v2/fizzbuzz?limit=100
func Logger(log *log.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			t := time.Now()
			defer func() {
				host, _, _ := net.SplitHostPort(r.RemoteAddr)
				if header := r.Header.Get("X-Forwarded-For"); header != "" {
					ips := strings.Split(header, ",")
					if len(ips) > 0 {
						firstIP := strings.TrimSpace(ips[0])
						validIP := net.ParseIP(firstIP) != nil
						if validIP {
							host = firstIP
						}
					}
				}
				secs := float64(time.Since(t)) / float64(time.Second)
				log.Println(host, fmt.Sprintf("%.6fs", secs), r.Method, r.URL)
			}()
			h.ServeHTTP(rw, r)
		})
	}
}
