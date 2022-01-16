package fizzbuzz

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func mustWrite(t *testing.T, f func(w io.Writer) error) string {
	t.Helper()
	var buf bytes.Buffer
	if err := f(&buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	i := strings.LastIndexByte(s, '\n')
	if i != len(s)-1 {
		t.Fatal("missing final newline")
	}
	if strings.LastIndexByte(s[:i], '\n') != -1 {
		t.Fatal("more than one final newline")
	}
	return s[:i] // trims final newline
}

func compare(t *testing.T, c Config) {
	t.Helper()
	b1 := mustWrite(t, c.WriteInto)
	b2 := mustWrite(t, c.WriteInto2)
	if b1 != b2 {
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
	compare(t, Config{0, 1, 1, "", ""})
	compare(t, Config{1, 1, 1, "", ""})
	compare(t, Config{1, 2, 3, "", ""})
	compare(t, Config{2, 2, 3, "", ""})
	compare(t, Config{3, 2, 3, "", ""})
	compare(t, Config{3, 2, 3, `"`, `"`})
}

func TestWriteInto(t *testing.T) {
	testCases := []struct {
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
	}
	for _, tc := range testCases {
		got := mustWrite(t, tc.input.WriteInto)
		if tc.expected != got {
			t.Errorf("WriteInto give an unexpected result with %#v\nexpected: %s\ngot:      %s", tc.input, tc.expected, got)
		}
	}
}

type closed struct{}

var errClosed = errors.New("writer is closed")

func (closed) Write([]byte) (n int, err error) { return 0, errClosed }

func TestWriteIntoClosedBuffer(t *testing.T) {
	if err := (&Config{0, 1, 1, "", ""}).WriteInto(closed{}); err != errClosed {
		t.Fatal("WriteInto should return the writer error", err)
	}
}

func benchmarkWriteInto(b *testing.B, limit int) {
	b.Helper()
	c := Config{limit, 2, 3, "fizz", "buzz"}
	for i := 0; i < b.N; i++ {
		if err := c.WriteInto(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkWriteInto2(b *testing.B, limit int) {
	b.Helper()
	c := Config{limit, 2, 3, "fizz", "buzz"}
	for i := 0; i < b.N; i++ {
		if err := c.WriteInto2(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

const (
	small  = 10
	medium = 1_000
	big    = 1_000_000
	huge   = 10_000_000
)

func BenchmarkWriteInto(b *testing.B) {
	b.Run("small", func(b *testing.B) { benchmarkWriteInto(b, small) })
	b.Run("medium", func(b *testing.B) { benchmarkWriteInto(b, medium) })
	b.Run("big", func(b *testing.B) { benchmarkWriteInto(b, big) })
	b.Run("huge", func(b *testing.B) { benchmarkWriteInto(b, huge) })
}

func BenchmarkWriteInto2(b *testing.B) {
	b.Run("small", func(b *testing.B) { benchmarkWriteInto2(b, small) })
	b.Run("medium", func(b *testing.B) { benchmarkWriteInto2(b, medium) })
	b.Run("big", func(b *testing.B) { benchmarkWriteInto2(b, big) })
	b.Run("huge", func(b *testing.B) { benchmarkWriteInto2(b, huge) })
}
