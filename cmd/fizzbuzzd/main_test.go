package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/xpetit/fizzbuzz/v5"
	main "github.com/xpetit/fizzbuzz/v5/cmd/fizzbuzzd"

	"golang.org/x/exp/slices"
)

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

const gotWant = "\ngot:  %+v\nwant: %+v"

func equal[T comparable](t *testing.T, descr string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s:"+gotWant, descr, got, want)
	}
}

func test(t *testing.T, c main.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start API
	runErr := make(chan error)
	go func() {
		runErr <- c.Run(ctx)
	}()

	client := http.Client{Timeout: time.Second}

	request := func(method, path string) (int, []byte, error) {
		req, err := http.NewRequest(method, "http://"+c.Addr+"/api/v2/"+path, nil)
		if err != nil {
			return 0, nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return 0, nil, err
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return 0, nil, err
		}
		if err := resp.Body.Close(); err != nil {
			return 0, nil, err
		}
		return resp.StatusCode, b, nil
	}

	for { // Wait for the HTTP server to be ready
		time.Sleep(100 * time.Millisecond)
		code, _, err := request("GET", "ready")
		if err != nil {
			continue
		}
		if code == http.StatusOK {
			break
		}
	}

	assertBadRequest := func(t *testing.T, method, path string) {
		t.Helper()
		code, b, err := request(method, path)
		check(t, err)
		clientError := 400 <= code && code <= 499
		equal(t, fmt.Sprintf("code %d is a client error", code), clientError, true)
		var resp map[string]string
		check(t, json.Unmarshal(b, &resp))
		_, ok := resp["error"]
		equal(t, `"error" field exists`, ok, true)
		equal(t, `"error" field is not empty`, resp["error"] != "", true)
	}

	// The endpoints only accept GET/HEAD HTTP methods
	invalidMethods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
	paths := []string{"fizzbuzz", "fizzbuzz/stats"}
	for _, path := range paths {
		for _, method := range invalidMethods {
			assertBadRequest(t, method, path)
		}
	}

	// fizzbuzz/stats endpoint doesn't accept query parameters
	assertBadRequest(t, "GET", "fizzbuzz/stats?unexpected_query")

	// fizzbuzz endpoint doesn't accept malformatted/unexpected/invalid query parameters
	invalidQueries := []string{
		"?unknown",
		"?limit=a",
		"?int1=a",
		"?int2=a",
		"?int1=0",
		"?int2=0",
		"?int1=-1",
		"?int2=-1",
		"?limit=",
		"?int1=",
		"?int2=",
		"?;",
	}
	for _, query := range invalidQueries {
		assertBadRequest(t, "GET", "fizzbuzz"+query)
	}

	assertStats := func(t *testing.T, count int, cfg fizzbuzz.Config) {
		t.Helper()
		var stats struct {
			MostFrequent struct {
				Config fizzbuzz.Config `json:"config"`
				Count  int             `json:"count"`
			} `json:"most_frequent"`
		}
		code, b, err := request("GET", "fizzbuzz/stats")
		check(t, err)
		equal(t, "HTTP code", code, http.StatusOK)
		check(t, json.Unmarshal(b, &stats))
		equal(t, "stats count", stats.MostFrequent.Count, count)
		equal(t, "stats config", stats.MostFrequent.Config, cfg)
	}

	getFizzbuzz := func(t *testing.T, cfg fizzbuzz.Config) (res []string) {
		t.Helper()
		code, b, err := request("GET", "fizzbuzz?"+url.Values{
			"str1":  {cfg.Str1},
			"str2":  {cfg.Str2},
			"limit": {strconv.Itoa(cfg.Limit)},
			"int1":  {strconv.Itoa(cfg.Int1)},
			"int2":  {strconv.Itoa(cfg.Int2)},
		}.Encode())
		check(t, err)
		equal(t, "HTTP code", code, http.StatusOK)
		check(t, json.Unmarshal(b, &res))
		return
	}

	assertStats(t, 0, fizzbuzz.Config{})

	baseConf := fizzbuzz.Config{
		Str1:  "fizz",
		Str2:  "buzz",
		Limit: 13,
		Int1:  3,
		Int2:  4,
	}
	getFizzbuzz(t, baseConf)
	assertStats(t, 1, baseConf)

	// When different Fizz buzz are called the same amount of times, only the "smallest" is returned
	// So incrementing Fizz buzz will not change the config returned by stats
	cfg := baseConf

	cfg.Limit++
	getFizzbuzz(t, cfg)
	assertStats(t, 1, baseConf)

	cfg.Int1++
	getFizzbuzz(t, cfg)
	assertStats(t, 1, baseConf)

	cfg.Int2++
	getFizzbuzz(t, cfg)
	assertStats(t, 1, baseConf)

	// On the contrary, "smaller" Fizz buzz configs update the stats
	cfg = baseConf

	cfg.Limit--
	getFizzbuzz(t, cfg)
	assertStats(t, 1, cfg)

	cfg.Int1--
	getFizzbuzz(t, cfg)
	assertStats(t, 1, cfg)

	cfg.Int2--
	getFizzbuzz(t, cfg)
	assertStats(t, 1, cfg)

	cfg.Str1 = "a"
	getFizzbuzz(t, cfg)
	assertStats(t, 1, cfg)

	cfg.Str2 = "a"
	getFizzbuzz(t, cfg)
	assertStats(t, 1, cfg)

	// Requesting the same Fizz buzz config increments the counter
	for i := 0; i < 5; i++ {
		getFizzbuzz(t, baseConf)
	}
	assertStats(t, 6, baseConf)

	// Fizz buzz works as intended
	tests := map[fizzbuzz.Config][]string{
		{Limit: -1, Str1: "", Str2: "", Int1: 1, Int2: 1}:         {},
		{Limit: -1, Str1: "a", Str2: "", Int1: 1, Int2: 1}:        {},
		{Limit: -1, Str1: "a", Str2: "a", Int1: 1, Int2: 1}:       {},
		{Limit: 0, Str1: "", Str2: "", Int1: 1, Int2: 1}:          {},
		{Limit: 0, Str1: "", Str2: "a", Int1: 1, Int2: 1}:         {},
		{Limit: 1, Str1: "", Str2: "", Int1: 1, Int2: 1}:          {""},
		{Limit: 1, Str1: "", Str2: "a", Int1: 1, Int2: 1}:         {"a"},
		{Limit: 1, Str1: "a", Str2: "", Int1: 1, Int2: 1}:         {"a"},
		{Limit: 1, Str1: "a", Str2: "b", Int1: 1, Int2: 1}:        {"ab"},
		{Limit: 1, Str1: "", Str2: "", Int1: 2, Int2: 2}:          {"1"},
		{Limit: 1, Str1: "", Str2: "", Int1: 2, Int2: 3}:          {"1"},
		{Limit: 1, Str1: `"`, Str2: "", Int1: 1, Int2: 1}:         {`"`},
		{Limit: 1, Str1: `👌🏻`, Str2: "", Int1: 1, Int2: 1}:        {`👌🏻`},
		{Limit: 2, Str1: "a", Str2: "b", Int1: 1, Int2: 2}:        {"a", "ab"},
		{Limit: 2, Str1: "a", Str2: "b", Int1: 2, Int2: 3}:        {"1", "a"},
		{Limit: 2, Str1: "a", Str2: "b", Int1: 3, Int2: 1}:        {"b", "b"},
		{Limit: 2, Str1: "a", Str2: "b", Int1: 3, Int2: 3}:        {"1", "2"},
		{Limit: 3, Str1: "a", Str2: "b", Int1: 3, Int2: 3}:        {"1", "2", "ab"},
		{Limit: 3, Str1: "a", Str2: "b", Int1: 3, Int2: 4}:        {"1", "2", "a"},
		{Limit: 4, Str1: "a", Str2: "b", Int1: 3, Int2: 4}:        {"1", "2", "a", "b"},
		{Limit: 6, Str1: "a", Str2: "b", Int1: 2, Int2: 3}:        {"1", "a", "b", "a", "5", "ab"},
		{Limit: 13, Str1: "fizz", Str2: "buzz", Int1: 3, Int2: 4}: {"1", "2", "fizz", "buzz", "5", "fizz", "7", "buzz", "fizz", "10", "11", "fizzbuzz", "13"},
	}
	for input, want := range tests {
		got := getFizzbuzz(t, input)
		if !slices.Equal(got, want) {
			t.Fatalf(gotWant, got, want)
		}
	}

	// Stop API
	cancel()
	check(t, <-runErr)
}

func TestMain(t *testing.T) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "60606"
	}
	addr := "127.0.0.1:" + port
	switch {
	case !t.Run("map", func(t *testing.T) {
		test(t, main.Config{Addr: addr, DBFile: "off"})
	}):
	case !t.Run("memory_DB", func(t *testing.T) {
		test(t, main.Config{Addr: addr, DBFile: ":memory:"})
	}):
	case !t.Run("file_DB", func(t *testing.T) {
		test(t, main.Config{Addr: addr, DBFile: filepath.Join(t.TempDir(), "data.db")})
	}):
	}
}
