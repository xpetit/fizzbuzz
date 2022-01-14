package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/xpetit/fizzbuzz"
)

// Stats holds a protected (thread safe) hit count
type Stats struct {
	sync.RWMutex
	m map[fizzbuzz.Config]int
}

// setInt sets the int pointed to by p to the value found in the values, or a default value.
// It returns an error if the value found in the values cannot be parsed as an int or
// if the resulting value is not strictly positive.
func setInt(values url.Values, key string, defaultValue int, p *int) error {
	if values.Has(key) {
		i, err := strconv.Atoi(values.Get(key))
		if err != nil {
			err := err.(*strconv.NumError)
			return fmt.Errorf("parsing %s %q: %w", key, err.Num, err.Err)
		}
		if i < 1 {
			return fmt.Errorf("setting %s to %d: must be strictly positive", key, i)
		}
		*p = i
	} else {
		*p = defaultValue
	}
	return nil
}

// setString sets the string pointed to by p to the value found in the values, or a default value.
func setString(values url.Values, key, value string, p *string) {
	if values.Has(key) {
		*p = values.Get(key)
	} else {
		*p = value
	}
}

// HandleFizzBuzz is an HTTP handler that answers with a JSON array containing the Fizz buzz values.
// It accepts optional URL query parameters to change the default config.
func (s *Stats) HandleFizzBuzz(rw http.ResponseWriter, r *http.Request) {
	// Parse query parameters with default values
	var c fizzbuzz.Config
	values := r.URL.Query()
	if err := setInt(values, "limit", 10, &c.Limit); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	} else if err := setInt(values, "int1", 2, &c.X); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	} else if err := setInt(values, "int2", 3, &c.Y); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	setString(values, "str1", "fizz", &c.A)
	setString(values, "str2", "buzz", &c.B)

	// Update fizzbuzz stats
	s.Lock()
	if s.m == nil {
		s.m = map[fizzbuzz.Config]int{}
	}
	s.m[c]++
	s.Unlock()

	// Write fizzbuzz
	c.WriteTo(rw)
}

// HandleStats is an HTTP handler that answers with a JSON object describing the most frequent Fizz buzz config
// If no subsequent call to fizzbuzz has been made, the "most_frequent" object is null.
func (s *Stats) HandleStats(rw http.ResponseWriter, r *http.Request) {
	var result struct {
		MostFrequent *fizzbuzz.Config `json:"most_frequent"`
		Count        int              `json:"count"`
	}
	s.RLock()
	for f, count := range s.m {
		if count > result.Count {
			result.Count = count
			f := f // shallow copy
			result.MostFrequent = &f
		}
	}
	s.RUnlock()
	json.NewEncoder(rw).Encode(result)
}
