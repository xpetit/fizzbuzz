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

type closed struct{}

var errClosed = errors.New("writer is closed")

func (closed) Write([]byte) (int, error) { return 0, errClosed }

func TestWriteInto(t *testing.T) {
	t.Parallel()

	// tests valid configurations
	t.Run("pass", func(t *testing.T) {
		t.Parallel()
		validCases := []struct {
			input    fizzbuzz.Config
			expected string
			name     string
		}{
			{fizzbuzz.Config{0, 1, 1, "", ""}, `[]`, "empty"},
			{fizzbuzz.Config{1, 1, 1, "", ""}, `[""]`, "str1"},
			{fizzbuzz.Config{1, 2, 2, "", ""}, `["1"]`, "number"},
			{fizzbuzz.Config{1, 1, 1, "", "a"}, `["a"]`, "str2"},
			{fizzbuzz.Config{1, 1, 1, "a", ""}, `["a"]`, "str1"},
			{fizzbuzz.Config{1, 1, 1, "a", "b"}, `["ab"]`, "str1str2"},
			{fizzbuzz.Config{2, 1, 2, "a", "b"}, `["a","ab"]`, "str1,str1str2"},
			{fizzbuzz.Config{2, 3, 1, "a", "b"}, `["b","b"]`, "str2,str2"},
			{fizzbuzz.Config{2, 2, 3, "a", "b"}, `["1","a"]`, "number,str1"},
			{fizzbuzz.Config{2, 3, 3, "a", "b"}, `["1","2"]`, "number,number"},
			{fizzbuzz.Config{1, 1, 1, `"`, ""}, `["\""]`, "str1_escaped)"},
			{fizzbuzz.Config{13, 3, 4, "fizz", "buzz"}, `["1","2","fizz","buzz","5","fizz","7","buzz","fizz","10","11","fizzbuzz","13"]`, "complete_suite"},
			{*fizzbuzz.Default(), `["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`, "default_suite"},
		}
		for _, vc := range validCases {
			vc := vc
			t.Run(vc.name, func(t *testing.T) {
				t.Parallel()
				got, err := write(t, vc.input.WriteInto)
				if err != nil {
					t.Fatal(err)
				}
				if vc.expected != got {
					t.Errorf("WriteInto give an unexpected result with %#v\nexpected: %s\ngot:      %s", vc.input, vc.expected, got)
				}
			})
		}
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		// tests invalid configurations
		t.Run("invalid", func(t *testing.T) {
			t.Parallel()
			invalidInputs := []fizzbuzz.Config{
				{-1, -1, -1, "", "negative_int1,negative_int2"},
				{1, -1, -1, "", "negative_int1,negative_int2"},
				{1, -1, 1, "", "negative_int1"},
				{1, 1, -1, "", "negative_int2"},
			}
			for _, ic := range invalidInputs {
				if _, err := write(t, ic.WriteInto); !errors.Is(err, fizzbuzz.ErrInvalidInput) {
					t.Fatalf("WriteInto should return an ErrInvalidInput with %#v", ic)
				}
			}
		})
		t.Run("closed", func(t *testing.T) {
			t.Parallel()
			if err := fizzbuzz.Default().WriteInto(closed{}); err != errClosed {
				t.Fatal("WriteInto should return the writer error", err)
			}
		})
	})

	// tests WriteInto and WriteInto2 side by side, reporting any inconsistencies
	t.Run("compare", func(t *testing.T) {
		t.Parallel()
		type testCase struct {
			input fizzbuzz.Config
			name  string
		}
		testCases := []testCase{
			{fizzbuzz.Config{-1, 1, 1, "", ""}, "empty"},
			{fizzbuzz.Config{-1, -1, -1, "", ""}, "negative_int1,negative_int2"},
			{fizzbuzz.Config{1, -1, -1, "", ""}, "negative_int1,negative_int2"},
			{fizzbuzz.Config{1, -1, 1, "", ""}, "negative_int1"},
			{fizzbuzz.Config{1, 1, -1, "", ""}, "negative_int2"},
			{fizzbuzz.Config{0, 1, 1, "", ""}, "empty"},
			{fizzbuzz.Config{1, 1, 1, "", ""}, "str1"},
			{fizzbuzz.Config{1, 2, 3, "", ""}, "number"},
			{fizzbuzz.Config{2, 2, 3, "", ""}, "number,str1"},
			{fizzbuzz.Config{3, 2, 3, "", ""}, "number,str1,str2"},
			{fizzbuzz.Config{3, 2, 3, `"`, `"`}, "number,str1_escaped,str2_escaped"},
		}
		for i := -10; i < 100; i++ {
			d := fizzbuzz.Default()
			d.Limit = i
			testCases = append(testCases, testCase{*d, fmt.Sprint("limit_", i)})
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// make sure that WriteInto and WriteInto2 behave in the same way
				b1, err1 := write(t, tc.input.WriteInto)
				b2, err2 := write(t, tc.input.WriteInto2)
				format := "WriteInto and WriteInto2 give different results with %#v\nWriteInto:  %s\nWriteInto2: %s"
				if err1 == nil && err2 != nil {
					t.Fatalf(format, tc.input, b1, err2)
				} else if err1 != nil && err2 == nil {
					t.Fatalf(format, tc.input, err1, b2)
				} else if err1 != nil && err2 != nil {
					if errors.Is(err1, fizzbuzz.ErrInvalidInput) != errors.Is(err2, fizzbuzz.ErrInvalidInput) {
						t.Fatalf(format, tc.input, err1, err2)
					}
				} else if b1 != b2 {
					t.Fatalf(format, tc.input, b1, b2)
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
