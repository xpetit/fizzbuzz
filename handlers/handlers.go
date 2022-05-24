package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/xpetit/fizzbuzz/v5"
)

type Stats interface {
	Increment(cfg fizzbuzz.Config) error
	MostFrequent() (count int, cfg fizzbuzz.Config, err error)
}

type handlers struct {
	stats Stats
}

// Fizzbuzz returns Fizz buzz HTTP handlers.
func Fizzbuzz(stats Stats) handlers {
	return handlers{
		stats: stats,
	}
}

// jsonErr is a helper function to respond a JSON-formatted error.
func jsonErr(rw http.ResponseWriter, message string, code int) {
	rw.WriteHeader(code)
	if err := json.NewEncoder(rw).Encode(struct {
		Error string `json:"error"`
	}{message}); err != nil {
		log.Println("write error:", err)
	}
}

// Handle is an HTTP handler that answers with a JSON array containing the Fizz buzz values.
// It accepts optional URL query parameters to change the default config.
func (fb handlers) Handle(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		jsonErr(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// parse query parameters with default values
	c := fizzbuzz.Default()
	values := r.URL.Query()

	intValues := map[string]*int{
		"int1":  &c.Int1,
		"int2":  &c.Int2,
		"limit": &c.Limit,
	}
	for key, target := range intValues {
		if values.Has(key) {
			i, err := strconv.Atoi(values.Get(key))
			if err != nil {
				err := err.(*strconv.NumError)
				jsonErr(rw, fmt.Sprintf("parsing %s %q: %s", key, err.Num, err.Err), http.StatusBadRequest)
				return
			}
			*target = i
		}
	}

	strValues := map[string]*string{
		"str1": &c.Str1,
		"str2": &c.Str2,
	}
	for key, target := range strValues {
		if values.Has(key) {
			*target = values.Get(key)
		}
	}

	// Write Fizz buzz and update the statistics in case of success
	if _, err := c.WriteTo(rw); err != nil {
		if errors.Is(err, fizzbuzz.ErrInvalidInput) {
			jsonErr(rw, err.Error(), http.StatusBadRequest)
		} else {
			log.Println("write error:", err)
		}
	} else if err := fb.stats.Increment(c); err != nil {
		log.Println("stats.increment:", err)
	}
}

// HandleStats is an HTTP handler that answers with a JSON object representing the most used Fizz buzz config.
// If no previous call to fizzbuzz has been made, most_frequent.count is 0 and most_frequent.config doesn't exist.
func (fb handlers) HandleStats(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		jsonErr(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if r.URL.RawQuery != "" {
		jsonErr(rw, "this endpoint takes no parameters", http.StatusBadRequest)
		return
	}

	count, cfg, err := fb.stats.MostFrequent()
	if err != nil {
		log.Println("stats.mostfrequent:", err)
		jsonErr(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var result struct {
		MostFrequent struct {
			Config *fizzbuzz.Config `json:"config,omitempty"`
			Count  int              `json:"count"`
		} `json:"most_frequent"`
	}
	if count > 0 {
		result.MostFrequent.Count = count
		result.MostFrequent.Config = &cfg
	}
	if err := json.NewEncoder(rw).Encode(result); err != nil {
		log.Println("write error:", err)
	}
}
