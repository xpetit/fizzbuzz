package fizzbuzz

import (
	"io"
	"testing"
)

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
