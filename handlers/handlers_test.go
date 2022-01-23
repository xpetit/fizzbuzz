package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/xpetit/fizzbuzz/v3/handlers"
)

// runParallel runs a parallel subtest.
func runParallel(t *testing.T, name string, f func(t *testing.T)) bool {
	return t.Run(name, func(t *testing.T) {
		t.Parallel()
		f(t)
	})
}

func resp(method, target string, handle http.HandlerFunc) *http.Response {
	req := httptest.NewRequest(method, "http://"+target, nil)
	rw := httptest.NewRecorder()
	handle(rw, req)
	return rw.Result()
}

func get(values url.Values, handle http.HandlerFunc) string {
	// It is safe to ignore the error:
	// The Response.Body is guaranteed to be non-nil and Body.Read call is
	// guaranteed to not return any error other than io.EOF.
	b, _ := io.ReadAll(resp("", "?"+values.Encode(), handle).Body)
	return strings.TrimRight(string(b), "\n")
}

func TestFizzbuzz(t *testing.T) {
	runParallel(t, "invalid method", func(t *testing.T) {
		invalidMethods := []string{
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}
		for _, method := range invalidMethods {
			method := method // capture range variable
			var fb handlers.Fizzbuzz
			handlerFuncs := map[string]http.HandlerFunc{
				"Handle":        fb.Handle,
				"HandleStats":   fb.HandleStats,
				"HandleStatsV1": fb.HandleStatsV1,
			}
			for name, handle := range handlerFuncs {
				name := name     // capture range variable
				handle := handle // capture range variable
				runParallel(t, method+" "+name, func(t *testing.T) {
					if code := resp(method, "", handle).StatusCode; code != http.StatusMethodNotAllowed {
						t.Error(method, name, "should not be allowed, received:", code, http.StatusText(code))
					}
				})
			}
		}
	})
	runParallel(t, "invalid query", func(t *testing.T) {
		invalidQueries := []string{
			"?limit=a",
			"?int1=a",
			"?int2=a",
			"?int1=0",
			"?int2=0",
			"?int1=-1",
			"?int2=-1",
		}
		for _, query := range invalidQueries {
			query := query // capture range variable
			runParallel(t, query, func(t *testing.T) {
				var fb handlers.Fizzbuzz
				if code := resp("", query, fb.Handle).StatusCode; code != http.StatusBadRequest {
					t.Error(query, "query should not be allowed, received:", code, http.StatusText(code))
				}
			})
		}
	})
	runParallel(t, "pass", func(t *testing.T) {
		var fb handlers.Fizzbuzz

		expected := `{"most_frequent":{"count":0}}`
		if got := get(nil, fb.HandleStats); got != expected {
			t.Fatal("unexpected output of HandleStats with no prior call to Handle, got:", got, "expected:", expected)
		}

		expected = `["1","a","b","a","5","ab"]`
		if got := get(url.Values{
			"limit": {"6"},
			"int1":  {"2"},
			"int2":  {"3"},
			"str1":  {"a"},
			"str2":  {"b"},
		}, fb.Handle); got != expected {
			t.Fatal("unexpected output of Handle, got:", got, "expected:", expected)
		}

		expected = `{"most_frequent":{"limit":6,"int1":2,"int2":3,"str1":"a","str2":"b"},"count":1}`
		if got := get(nil, fb.HandleStatsV1); got != expected {
			t.Fatal("unexpected output of HandleStatsV1 after a call to Handle, got:", got, "expected:", expected)
		}

		expected = `{"most_frequent":{"count":1,"config":{"limit":6,"int1":2,"int2":3,"str1":"a","str2":"b"}}}`
		if got := get(nil, fb.HandleStats); got != expected {
			t.Fatal("unexpected output of HandleStats after a call to Handle, got:", got, "expected:", expected)
		}

		// Add a second hit count of 1 for a different config
		get(nil, fb.Handle)

		// Make sure that HandleStats always return the same output for Fizzbuzz configs with the same count
		for i := 0; i < 100; i++ {
			i := i // capture range variable
			runParallel(t, strconv.Itoa(i), func(t *testing.T) {
				expected := `{"most_frequent":{"count":1,"config":{"limit":6,"int1":2,"int2":3,"str1":"a","str2":"b"}}}`
				if got := get(nil, fb.HandleStats); got != expected {
					t.Fatal("unexpected output of HandleStats after a call to Handle, got:", got, "expected:", expected)
				}
			})
		}
	})
}
