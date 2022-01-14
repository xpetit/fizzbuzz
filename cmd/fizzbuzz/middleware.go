package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func logger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			t := time.Now()
			defer func() {
				ip := r.RemoteAddr[:strings.LastIndexByte(r.RemoteAddr, ':')]
				secs := float64(time.Since(t)) / float64(time.Second)
				logger.Println(ip, fmt.Sprintf("%.6fs", secs), r.Method, r.URL)
			}()
			h.ServeHTTP(rw, r)
		})
	}
}
