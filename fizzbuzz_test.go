package fizzbuzz_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/xpetit/fizzbuzz/v4"
)

func Example() {
	fizzbuzz.Default().WriteTo(os.Stdout)
	(&fizzbuzz.Config{Limit: 15, Int1: 3, Int2: 5, Str1: "a", Str2: "b"}).WriteTo(os.Stdout)
	// Output:
	// ["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
	// ["1","2","a","4","b","a","7","8","a","b","11","a","13","14","ab"]
}

// output is a testing helper function that calls f and returns:
// - what it wrote in a trimmed string
// - an error if one occured
// It also asserts that there is one and only one final newline.
func output(t *testing.T, c fizzbuzz.Config) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		return "", err
	}
	s := buf.String()
	i := strings.LastIndexByte(s, '\n')
	if i == -1 || i != len(s)-1 {
		t.Fatal("missing final newline")
	} else if strings.LastIndexByte(s[:i], '\n') != -1 {
		t.Fatal("more than one final newline")
	}
	return s[:i], nil // trims final newline
}

// closed implements io.Writer and always returns a errClosed error.
type closed struct{}

var errClosed = errors.New("writer is closed")

func (closed) Write([]byte) (int, error) { return 0, errClosed }

// runParallel runs a parallel subtest.
func runParallel(t *testing.T, name string, f func(t *testing.T)) bool {
	return t.Run(name, func(t *testing.T) {
		t.Parallel()
		f(t)
	})
}

func format(c fizzbuzz.Config) string {
	b, _ := json.Marshal(c)
	return string(b)
}

// TestWriteTo tests the WriteTo function with valid and invalid configurations.
// It uses the subtests introduced in Go 1.7, which gives fine-grained control over which test(s) to run.
//
// Run the comparison of the generated test case with a limit of 9:
// go test -run '/compare/limit:9\b'  # Note the word-boundary regex anchor '\b' to avoid matching "limit:91"
//
// Run the tests where WriteTo should fail and has a negative int2:
// go test -run /fail/int2:-
//
// Run the tests where WriteTo shouldn't write any value:
// go test -run //limit:[0-]{1}
//
// For more information about subtests and sub-benchmarks, please visit https://go.dev/blog/subtests
func TestWriteTo(t *testing.T) {
	type testCase struct {
		input    fizzbuzz.Config
		expected string
	}
	validCases := []testCase{
		{fizzbuzz.Config{-1, 1, 1, "", ""}, `[]`},
		{fizzbuzz.Config{-1, 1, 1, "a", ""}, `[]`},
		{fizzbuzz.Config{-1, 1, 1, "a", "a"}, `[]`},
		{fizzbuzz.Config{0, 1, 1, "", ""}, `[]`},
		{fizzbuzz.Config{0, 1, 1, "", "a"}, `[]`},
		{fizzbuzz.Config{1, 1, 1, "", ""}, `[""]`},
		{fizzbuzz.Config{1, 1, 1, "", "a"}, `["a"]`},
		{fizzbuzz.Config{1, 1, 1, "a", ""}, `["a"]`},
		{fizzbuzz.Config{1, 1, 1, "a", "b"}, `["ab"]`},
		{fizzbuzz.Config{1, 2, 2, "", ""}, `["1"]`},
		{fizzbuzz.Config{1, 2, 3, "", ""}, `["1"]`},
		{fizzbuzz.Config{2, 1, 2, "a", "b"}, `["a","ab"]`},
		{fizzbuzz.Config{2, 2, 3, "a", "b"}, `["1","a"]`},
		{fizzbuzz.Config{2, 3, 1, "a", "b"}, `["b","b"]`},
		{fizzbuzz.Config{2, 3, 3, "a", "b"}, `["1","2"]`},
		{fizzbuzz.Config{3, 3, 3, "a", "b"}, `["1","2","ab"]`},
		{fizzbuzz.Config{3, 3, 4, "a", "b"}, `["1","2","a"]`},
		{fizzbuzz.Config{4, 3, 4, "a", "b"}, `["1","2","a","b"]`},
		{fizzbuzz.Config{6, 2, 3, "a", "b"}, `["1","a","b","a","5","ab"]`},
		{fizzbuzz.Config{1, 1, 1, `"`, ""}, `["\""]`},
		{fizzbuzz.Config{13, 3, 4, "fizz", "buzz"}, `["1","2","fizz","buzz","5","fizz","7","buzz","fizz","10","11","fizzbuzz","13"]`},
		{*fizzbuzz.Default(), `["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`},
	}
	invalidCases := []testCase{
		{input: fizzbuzz.Config{-1, -1, -1, "", ""}},
		{input: fizzbuzz.Config{1, -1, -1, "", ""}},
		{input: fizzbuzz.Config{1, -1, 1, "", ""}},
		{input: fizzbuzz.Config{1, 1, -1, "", ""}},
	}

	// tests valid cases (WriteTo should not return an error)
	runParallel(t, "pass", func(t *testing.T) {
		for _, tc := range validCases {
			tc := tc // capture range variable
			runParallel(t, format(tc.input), func(t *testing.T) {
				if got, err := output(t, tc.input); err != nil {
					t.Fatal("WriteTo failed:", err)
				} else if tc.expected != got {
					t.Errorf("expected: %s, got: %s", tc.expected, got)
				}
			})
		}
	})

	// tests invalid cases (WriteTo should return an error)
	runParallel(t, "fail", func(t *testing.T) {
		runParallel(t, "closed", func(t *testing.T) {
			if _, err := fizzbuzz.Default().WriteTo(closed{}); err != errClosed {
				t.Error("WriteTo should return the writer error, instead it returned:", err)
			}
		})

		// tests invalid configurations
		for _, tc := range invalidCases {
			tc := tc // capture range variable
			runParallel(t, format(tc.input), func(t *testing.T) {
				if _, err := output(t, tc.input); !errors.Is(err, fizzbuzz.ErrInvalidInput) {
					t.Error("WriteTo should return an ErrInvalidInput")
				}
			})
		}
	})
}

// BenchmarkWriteTo benchmarks WriteTo with a default config and a limit of n
func BenchmarkWriteTo(b *testing.B) {
	c := fizzbuzz.Default()
	for n := -1; n <= 7; n++ {
		c.Limit = int(math.Pow10(n))
		b.Run(fmt.Sprintf("[limit:%.e]", float64(c.Limit)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if n, err := c.WriteTo(io.Discard); err != nil {
					b.Fatal(err)
				} else {
					b.SetBytes(n)
				}
			}
		})
	}
}
