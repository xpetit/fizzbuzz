package fizzbuzz_test

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/xpetit/fizzbuzz/v3"
)

func Example() {
	fizzbuzz.Default().WriteInto(os.Stdout)
	(&fizzbuzz.Config{Limit: 15, Int1: 3, Int2: 5, Str1: "a", Str2: "b"}).WriteInto(os.Stdout)
	// Output:
	// ["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
	// ["1","2","a","4","b","a","7","8","a","b","11","a","13","14","ab"]
}

// write is a testing helper function that calls f and returns:
// - what it wrote in a trimmed string
// - an error if one occured
// It also asserts that there is one and only one final newline.
func write(t *testing.T, f func(w io.Writer) error) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	if err := f(&buf); err != nil {
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

// TestWriteInto tests the WriteInto function with valid, invalid configurations and compares it to WriteInto2.
// It uses the subtests introduced in Go 1.7, which gives fine-grained control over which test(s) to run.
//
// Run the comparison of the generated test case with a limit of 9:
// go test -run '/compare/limit:9\b'  # Note the word-boundary regex anchor '\b' to avoid matching "limit:91"
//
// Run the tests where WriteInto should fail and has a negative int2:
// go test -run /fail/int2:-
//
// Run the tests where WriteInto shouldn't write any value:
// go test -run //limit:[0-]{1}
//
// For more information about subtests and sub-benchmarks, please visit https://go.dev/blog/subtests
func TestWriteInto(t *testing.T) {
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
	allTestCases := append(validCases, invalidCases...)

	// tests valid cases (WriteInto should not return an error)
	runParallel(t, "pass", func(t *testing.T) {
		for _, tc := range validCases {
			tc := tc // capture range variable
			runParallel(t, tc.input.String(), func(t *testing.T) {
				if got, err := write(t, tc.input.WriteInto); err != nil {
					t.Fatal("WriteInto failed:", err)
				} else if tc.expected != got {
					t.Errorf("expected: %s, got: %s", tc.expected, got)
				}
			})
		}
	})

	// tests invalid cases (WriteInto should return an error)
	runParallel(t, "fail", func(t *testing.T) {
		runParallel(t, "closed", func(t *testing.T) {
			if err := fizzbuzz.Default().WriteInto(closed{}); err != errClosed {
				t.Error("WriteInto should return the writer error, instead it returned:", err)
			}
		})

		// tests invalid configurations
		for _, tc := range invalidCases {
			tc := tc // capture range variable
			runParallel(t, tc.input.String(), func(t *testing.T) {
				if _, err := write(t, tc.input.WriteInto); !errors.Is(err, fizzbuzz.ErrInvalidInput) {
					t.Error("WriteInto should return an ErrInvalidInput")
				}
			})
		}
	})

	// tests WriteInto and WriteInto2 side by side, reporting any inconsistencies
	runParallel(t, "compare", func(t *testing.T) {
		if testing.Short() {
			t.SkipNow()
		}

		testCases := append([]testCase(nil), allTestCases...) // copy all test cases

		// add valid test cases with a variable limit
		for limit := -10; limit < 100; limit++ {
			testCases = append(testCases, testCase{input: *withLimit(limit)})
		}

		// add random test cases
		ss := []string{
			``,
			`a`,
			`"`,
			`"""""`,
			"\u2063", // Invisible Separator
			`abc`,
			`1`,
			`2`,
			`12`,
		}
		randStr := func() (s string) {
			for i := 0; i < rand.Intn(4); i++ {
				s += ss[rand.Intn(len(ss))] // pick a random string from ss
			}
			return // the concatenation of 0 to 3 random strings from ss
		}
		for i := 0; i < 100; i++ {
			testCases = append(testCases, testCase{input: fizzbuzz.Config{
				Limit: rand.Intn(100) - 10, // between -10 and 89
				Int1:  rand.Intn(100) - 10, // between -10 and 89
				Int2:  rand.Intn(100) - 10, // between -10 and 89
				Str1:  randStr(),
				Str2:  randStr(),
			}})
		}

		for _, tc := range testCases {
			tc := tc // capture range variable
			runParallel(t, tc.input.String(), func(t *testing.T) {
				// make sure that WriteInto and WriteInto2 behave in the same way
				b1, err1 := write(t, tc.input.WriteInto)
				b2, err2 := write(t, tc.input.WriteInto2)
				if err1 == nil && err2 != nil {
					t.Errorf("WriteInto: %s, WriteInto2: %s", b1, err2)
				} else if err1 != nil && err2 == nil {
					t.Errorf("WriteInto: %s, WriteInto2: %s", err1, b2)
				} else if err1 != nil && err2 != nil {
					if errors.Is(err1, fizzbuzz.ErrInvalidInput) != errors.Is(err2, fizzbuzz.ErrInvalidInput) {
						t.Errorf("WriteInto: %s, WriteInto2: %s", err1, err2)
					}
				} else if b1 != b2 { // err1 == nil && err2 == nil
					t.Errorf("WriteInto: %s, WriteInto2: %s", b1, b2)
				}
			})
		}
	})
}

func withLimit(limit int) *fizzbuzz.Config {
	d := fizzbuzz.Default()
	d.Limit = limit
	return d
}

var (
	small  = withLimit(10)
	medium = withLimit(1_000)
	big    = withLimit(1_000_000)
	huge   = withLimit(10_000_000)
)

// discard benchmarks f with io.Discard writer
func discard(b *testing.B, f func(w io.Writer) error) {
	b.Helper()
	for i := 0; i < b.N; i++ {
		if err := f(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteInto(b *testing.B) {
	b.Run("small", func(b *testing.B) { discard(b, small.WriteInto) })
	b.Run("medium", func(b *testing.B) { discard(b, medium.WriteInto) })
	b.Run("big", func(b *testing.B) { discard(b, big.WriteInto) })
	b.Run("huge", func(b *testing.B) { discard(b, huge.WriteInto) })
}

func BenchmarkWriteInto2(b *testing.B) {
	b.Run("small", func(b *testing.B) { discard(b, small.WriteInto2) })
	b.Run("medium", func(b *testing.B) { discard(b, medium.WriteInto2) })
	b.Run("big", func(b *testing.B) { discard(b, big.WriteInto2) })
	b.Run("huge", func(b *testing.B) { discard(b, huge.WriteInto2) })
}
