package fizzbuzz_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/xpetit/fizzbuzz"
)

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
	if i != len(s)-1 {
		t.Fatal("missing final newline")
	}
	if strings.LastIndexByte(s[:i], '\n') != -1 {
		t.Fatal("more than one final newline")
	}
	return s[:i], nil // trims final newline
}

// closed implements io.Writer and always returns a errClosed error.
type closed struct{}

var errClosed = errors.New("writer is closed")

func (closed) Write([]byte) (int, error) { return 0, errClosed }

// test runs a parallel subtest
func test(t *testing.T, name string, f func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		f(t)
	})
}

// TestWriteInto tests the WriteInto function with valid, invalid configurations and compares it to WriteInto2.
// It uses the subtests introduced in Go 1.7, which gives fine-grained control over which test(s) to run.
//
// Run the comparision of the generated test case with a limit of 9:
// go test -run /compare/limit_9$  # Note the EOL regex symbol '$' to avoid matching "limit_90"
//
// Run the test where WriteInto should fail because of a negative int2:
// go test -run /fail/^negative_int2$
//
// Run the tests where WriteInto should fail and has a negative int2:
// go test -run /fail/negative_int2
//
// Run the test where WriteInto shouldn't write any value:
// go test -run //empty
//
// For more information about subtests and sub-benchmarks, please visit https://go.dev/blog/subtests
func TestWriteInto(t *testing.T) {
	type testCase struct {
		input    fizzbuzz.Config
		expected string
		name     string
	}

	// tests valid configurations
	test(t, "pass", func(t *testing.T) {
		validCases := []testCase{
			{fizzbuzz.Config{0, 1, 1, "", "a"}, `[]`, "empty"},
			{fizzbuzz.Config{1, 1, 1, "", ""}, `[""]`, "str1str2"},
			{fizzbuzz.Config{1, 2, 2, "", ""}, `["1"]`, "number"},
			{fizzbuzz.Config{1, 1, 1, "", "a"}, `["a"]`, "str1str2"},
			{fizzbuzz.Config{1, 1, 1, "a", ""}, `["a"]`, "str1str2"},
			{fizzbuzz.Config{1, 1, 1, "a", "b"}, `["ab"]`, "str1str2"},
			{fizzbuzz.Config{2, 1, 2, "a", "b"}, `["a","ab"]`, "str1,str1str2"},
			{fizzbuzz.Config{2, 3, 1, "a", "b"}, `["b","b"]`, "str2,str2"},
			{fizzbuzz.Config{2, 2, 3, "a", "b"}, `["1","a"]`, "number,str1"},
			{fizzbuzz.Config{2, 3, 3, "a", "b"}, `["1","2"]`, "number,number"},
			{fizzbuzz.Config{1, 1, 1, `"`, ""}, `["\""]`, "str1_escaped)"},
			{fizzbuzz.Config{13, 3, 4, "fizz", "buzz"}, `["1","2","fizz","buzz","5","fizz","7","buzz","fizz","10","11","fizzbuzz","13"]`, "complete_suite"},
			{*fizzbuzz.Default(), `["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`, "default_suite"},
		}
		for _, tc := range validCases {
			tc := tc
			test(t, tc.name, func(t *testing.T) {
				got, err := write(t, tc.input.WriteInto)
				if err != nil {
					t.Fatalf("WriteInto failed with the valid test case %#v, err: %v", tc.input, err)
				}
				if tc.expected != got {
					t.Errorf("WriteInto give an unexpected result with %#v\nexpected: %s\ngot:      %s", tc.input, tc.expected, got)
				}
			})
		}
	})

	test(t, "fail", func(t *testing.T) {
		// tests invalid configurations
		invalidCases := []testCase{
			{fizzbuzz.Config{-1, -1, -1, "", ""}, "", "negative_int1,negative_int2"},
			{fizzbuzz.Config{1, -1, -1, "", ""}, "", "negative_int1,negative_int2"},
			{fizzbuzz.Config{1, -1, 1, "", ""}, "", "negative_int1"},
			{fizzbuzz.Config{1, 1, -1, "", ""}, "", "negative_int2"},
		}
		for _, tc := range invalidCases {
			tc := tc
			test(t, tc.name, func(t *testing.T) {
				if _, err := write(t, tc.input.WriteInto); !errors.Is(err, fizzbuzz.ErrInvalidInput) {
					t.Errorf("WriteInto should return an ErrInvalidInput with %#v", tc.input)
				}
			})
		}
		test(t, "closed", func(t *testing.T) {
			if err := fizzbuzz.Default().WriteInto(closed{}); err != errClosed {
				t.Error("WriteInto should return the writer error, instead it returned", err)
			}
		})
	})

	// tests WriteInto and WriteInto2 side by side, reporting any inconsistencies
	test(t, "compare", func(t *testing.T) {
		testCases := []testCase{
			{fizzbuzz.Config{-1, 1, 1, "", ""}, "", "empty"},
			{fizzbuzz.Config{-1, -1, -1, "", ""}, "", "negative_int1,negative_int2"},
			{fizzbuzz.Config{1, -1, -1, "", ""}, "", "negative_int1,negative_int2"},
			{fizzbuzz.Config{1, -1, 1, "", ""}, "", "negative_int1"},
			{fizzbuzz.Config{1, 1, -1, "", ""}, "", "negative_int2"},
			{fizzbuzz.Config{0, 1, 1, "", ""}, "", "empty"},
			{fizzbuzz.Config{1, 1, 1, "", ""}, "", "str1"},
			{fizzbuzz.Config{1, 2, 3, "", ""}, "", "number"},
			{fizzbuzz.Config{2, 2, 3, "", ""}, "", "number,str1"},
			{fizzbuzz.Config{3, 2, 3, "", ""}, "", "number,str1,str2"},
			{fizzbuzz.Config{3, 2, 3, `"`, `"`}, "", "number,str1_escaped,str2_escaped"},
		}
		for limit := -10; limit < 100; limit++ {
			d := fizzbuzz.Default()
			d.Limit = limit
			testCases = append(testCases, testCase{*d, "", fmt.Sprint("limit_", limit)})
		}
		for _, tc := range testCases {
			tc := tc
			test(t, tc.name, func(t *testing.T) {
				// make sure that WriteInto and WriteInto2 behave in the same way
				b1, err1 := write(t, tc.input.WriteInto)
				b2, err2 := write(t, tc.input.WriteInto2)
				format := "WriteInto and WriteInto2 give different results with %#v\nWriteInto:  %s\nWriteInto2: %s"
				if err1 == nil && err2 != nil {
					t.Errorf(format, tc.input, b1, err2)
				} else if err1 != nil && err2 == nil {
					t.Errorf(format, tc.input, err1, b2)
				} else if err1 != nil && err2 != nil {
					if errors.Is(err1, fizzbuzz.ErrInvalidInput) != errors.Is(err2, fizzbuzz.ErrInvalidInput) {
						t.Errorf(format, tc.input, err1, err2)
					}
				} else if b1 != b2 {
					t.Errorf(format, tc.input, b1, b2)
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
