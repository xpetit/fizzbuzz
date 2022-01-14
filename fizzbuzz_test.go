package fizzbuzz

import (
	"encoding/json"
	"io"
	"testing"
)

const (
	small  = 10
	medium = 1_000
	big    = 1_000_000
	huge   = 100_000_000
)

func writeTo(b *testing.B, limit int) {
	c := Config{limit, 2, 3, "fizz", "buzz"}
	for i := 0; i < b.N; i++ {
		if _, err := c.WriteTo(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func encode(b *testing.B, limit int) {
	c := Config{limit, 2, 3, "fizz", "buzz"}
	for i := 0; i < b.N; i++ {
		if err := json.NewEncoder(io.Discard).Encode(c.ToSlice()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteTo(b *testing.B) {
	b.Run("small", func(b *testing.B) { writeTo(b, small) })
	b.Run("medium", func(b *testing.B) { writeTo(b, medium) })
	b.Run("big", func(b *testing.B) { writeTo(b, big) })
	b.Run("huge", func(b *testing.B) { writeTo(b, huge) })
}

func BenchmarkEncode(b *testing.B) {
	b.Run("small", func(b *testing.B) { encode(b, small) })
	b.Run("medium", func(b *testing.B) { encode(b, medium) })
	b.Run("big", func(b *testing.B) { encode(b, big) })
	b.Run("huge", func(b *testing.B) { encode(b, huge) })
}
