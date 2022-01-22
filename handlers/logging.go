package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Logger is a basic HTTP middleware that logs queries with the following format:
// UTC with microsecond resolution, remote IP address, time taken, HTTP method, URL.
// Example:
//   "2022/01/14 15:48:50.301103 172.17.0.1 0.000003s GET /api/v2/fizzbuzz?limit=100"
func Logger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			t := time.Now()
			defer func() {
				// the colon is guaranteed to be found so the index is always positive
				ip := r.RemoteAddr[:strings.LastIndexByte(r.RemoteAddr, ':')]

				secs := float64(time.Since(t)) / float64(time.Second)
				logger.Println(ip, fmt.Sprintf("%.6fs", secs), r.Method, r.URL)
			}()
			h.ServeHTTP(rw, r)
		})
	}
}
