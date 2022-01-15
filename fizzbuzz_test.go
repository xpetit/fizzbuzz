package fizzbuzz

import (
	"io"
	"testing"
)

const (
	small  = 10
	medium = 1_000
	big    = 1_000_000
	huge   = 10_000_000
)

func writeWith(b *testing.B, limit int) {
	c := Config{limit, 2, 3, "fizz", "buzz"}
	for i := 0; i < b.N; i++ {
		if err := c.WriteWith(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func writeWith2(b *testing.B, limit int) {
	c := Config{limit, 2, 3, "fizz", "buzz"}
	for i := 0; i < b.N; i++ {
		if err := c.WriteWith2(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteWith(b *testing.B) {
	b.Run("small", func(b *testing.B) { writeWith(b, small) })
	b.Run("medium", func(b *testing.B) { writeWith(b, medium) })
	b.Run("big", func(b *testing.B) { writeWith(b, big) })
	b.Run("huge", func(b *testing.B) { writeWith(b, huge) })
}

func BenchmarkWriteWith2(b *testing.B) {
	b.Run("small", func(b *testing.B) { writeWith2(b, small) })
	b.Run("medium", func(b *testing.B) { writeWith2(b, medium) })
	b.Run("big", func(b *testing.B) { writeWith2(b, big) })
	b.Run("huge", func(b *testing.B) { writeWith2(b, huge) })
}
