package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/xpetit/fizzbuzz/v5"
)

// Fizzbuzz is an HTTP handler that serves a Fizzbuzz endpoint with statistics.
type Fizzbuzz struct {
	Stats interface {
		Increment(cfg *fizzbuzz.Config) error
		MostFrequent() (count int, cfg *fizzbuzz.Config, err error)
	}
}

// setInt sets the int pointed to by p to the value found in the values, or a default value.
// It returns an error if the value found in the values cannot be parsed as an int.
func setInt(p *int, values url.Values, key string, defaultValue int) error {
	if values.Has(key) {
		i, err := strconv.Atoi(values.Get(key))
		if err != nil {
			err := err.(*strconv.NumError)
			return fmt.Errorf("parsing %s %q: %w", key, err.Num, err.Err)
		}
		*p = i
	} else {
		*p = defaultValue
	}
	return nil
}

// setString sets the string pointed to by p to the value found in the values, or a default value.
func setString(p *string, values url.Values, key, value string) {
	if values.Has(key) {
		*p = values.Get(key)
	} else {
		*p = value
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
func (fb *Fizzbuzz) Handle(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		jsonErr(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters with default values
	var c fizzbuzz.Config
	values := r.URL.Query()
	if err := setInt(&c.Limit, values, "limit", 10); err != nil {
		jsonErr(rw, err.Error(), http.StatusBadRequest)
		return
	} else if err := setInt(&c.Int1, values, "int1", 2); err != nil {
		jsonErr(rw, err.Error(), http.StatusBadRequest)
		return
	} else if err := setInt(&c.Int2, values, "int2", 3); err != nil {
		jsonErr(rw, err.Error(), http.StatusBadRequest)
		return
	}
	setString(&c.Str1, values, "str1", "fizz")
	setString(&c.Str2, values, "str2", "buzz")

	// Write Fizz buzz and update the statistics in case of success
	if _, err := c.WriteTo(rw); err != nil {
		if errors.Is(err, fizzbuzz.ErrInvalidInput) {
			jsonErr(rw, err.Error(), http.StatusBadRequest)
		} else {
			log.Println("write error:", err)
		}
	} else if err := fb.Stats.Increment(&c); err != nil {
		log.Println("stats.increment:", err)
	}
}

// HandleStats is an HTTP handler that answers with a JSON object representing the most used Fizz buzz config.
// If no previous call to fizzbuzz has been made, most_frequent.count is 0 and most_frequent.config doesn't exist.
func (fb *Fizzbuzz) HandleStats(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		jsonErr(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	} else if len(r.URL.RawQuery) > 0 {
		jsonErr(rw, "this endpoint takes no parameters", http.StatusBadRequest)
		return
	}

	count, cfg, err := fb.Stats.MostFrequent()
	if err != nil {
		log.Println("stats.mostfrequent:", err)
		jsonErr(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var result struct {
		MostFrequent struct {
			Count  int              `json:"count"`
			Config *fizzbuzz.Config `json:"config,omitempty"`
		} `json:"most_frequent"`
	}
	result.MostFrequent.Count, result.MostFrequent.Config = count, cfg
	if err := json.NewEncoder(rw).Encode(result); err != nil {
		log.Println("write error:", err)
	}
}
