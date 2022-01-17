package fizzbuzz

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// write is testing helper function that calls f and returns:
// - what it wrote in a trimmed string
// - an error if one occured
// It also checks that there is one and only one final newline.
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

// mustWrite asserts that write worked
func mustWrite(t *testing.T, f func(w io.Writer) error) string {
	t.Helper()
	s, err := write(t, f)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func compare(t *testing.T, c Config) {
	t.Helper()
	b1, err1 := write(t, c.WriteInto)
	b2, err2 := write(t, c.WriteInto2)

	if err1 == nil && err2 != nil {
		t.Fatalf("WriteInto and WriteInto2 give different results with %#v\nWriteInto:  %s\nWriteInto2: %s", c, b1, err2)
	} else if err1 != nil && err2 == nil {
		t.Fatalf("WriteInto and WriteInto2 give different results with %#v\nWriteInto:  %s\nWriteInto2: %s", c, err1, b2)
	} else if err1 != nil && err2 != nil {
		if errors.Is(err1, ErrInvalidInput) != errors.Is(err2, ErrInvalidInput) {
			t.Fatalf("WriteInto and WriteInto2 give different results with %#v\nWriteInto:  %s\nWriteInto2: %s", c, err1, err2)
		}
	} else if b1 != b2 {
		t.Fatalf("WriteInto and WriteInto2 give different results with %#v\nWriteInto:  %s\nWriteInto2: %s", c, b1, b2)
	}
}

func TestWriteIntoCompare(t *testing.T) {
	for i := -10; i < 100; i++ {
		i := i
		t.Run(fmt.Sprintln("simple case with limit:", i), func(t *testing.T) {
			t.Parallel()
			compare(t, Config{i, 2, 3, "fizz", "buzz"})
		})
	}
	compare(t, Config{-1, 1, 1, "", ""})
	compare(t, Config{-1, -1, -1, "", ""})
	compare(t, Config{1, -1, -1, "", ""})
	compare(t, Config{1, -1, 1, "", ""})
	compare(t, Config{1, 1, -1, "", ""})
	compare(t, Config{0, 1, 1, "", ""})
	compare(t, Config{1, 1, 1, "", ""})
	compare(t, Config{1, 2, 3, "", ""})
	compare(t, Config{2, 2, 3, "", ""})
	compare(t, Config{3, 2, 3, "", ""})
	compare(t, Config{3, 2, 3, `"`, `"`})
}

func TestWriteInto(t *testing.T) {
	validCases := []struct {
		input    Config
		expected string
	}{
		{Config{0, 1, 1, "", ""}, `[]`},
		{Config{1, 1, 1, "", ""}, `[""]`},
		{Config{1, 2, 2, "", ""}, `["1"]`},
		{Config{1, 1, 1, "", "a"}, `["a"]`},
		{Config{1, 1, 1, "a", ""}, `["a"]`},
		{Config{1, 1, 1, "a", "b"}, `["ab"]`},
		{Config{2, 1, 2, "a", "b"}, `["a","ab"]`},
		{Config{2, 3, 1, "a", "b"}, `["b","b"]`},
		{Config{2, 2, 3, "a", "b"}, `["1","a"]`},
		{Config{2, 3, 3, "a", "b"}, `["1","2"]`},
		{Config{1, 1, 1, `"`, ""}, `["\""]`},
		{Config{13, 3, 4, "fizz", "buzz"}, `["1","2","fizz","buzz","5","fizz","7","buzz","fizz","10","11","fizzbuzz","13"]`},
		{*Default(), `["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`},
	}
	for _, vc := range validCases {
		got := mustWrite(t, vc.input.WriteInto)
		if vc.expected != got {
			t.Errorf("WriteInto give an unexpected result with %#v\nexpected: %s\ngot:      %s", vc.input, vc.expected, got)
		}
	}
}

type closed struct{}

var errClosed = errors.New("writer is closed")

func (closed) Write([]byte) (int, error) { return 0, errClosed }

func TestWriteIntoClosedBuffer(t *testing.T) {
	if err := Default().WriteInto(closed{}); err != errClosed {
		t.Fatal("WriteInto should return the writer error", err)
	}
}

func withLimit(limit int) *Config {
	d := Default()
	d.Limit = limit
	return d
}

var (
	small  = withLimit(10)
	medium = withLimit(1_000)
	big    = withLimit(1_000_000)
	huge   = withLimit(10_000_000)
)

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
