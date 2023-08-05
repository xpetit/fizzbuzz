package stats_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/xpetit/fizzbuzz/v5"
	"github.com/xpetit/fizzbuzz/v5/stats"
)

func Benchmark(b *testing.B) {
	randomFB := make([]fizzbuzz.Config, 0, 1_000_000)

	fbSet := map[fizzbuzz.Config]struct{}{}
	for len(randomFB) < cap(randomFB) {
		fb := fizzbuzz.Config{
			Int1:  rand.Intn(10_000),
			Int2:  rand.Intn(10_000),
			Limit: rand.Intn(10_000),
		}
		if _, ok := fbSet[fb]; !ok {
			fbSet[fb] = struct{}{}
			randomFB = append(randomFB, fb)
		}
	}

	mem := stats.Memory()
	b.Run("Memory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := mem.Increment(randomFB[rand.Intn(len(randomFB))]); err != nil {
				b.Fatal(err)
			}
		}
	})

	db, err := stats.OpenDB(context.Background(), ":memory:")
	if err != nil {
		b.Fatal("failed to open database:", err)
	}
	b.Run("DB", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := db.Increment(randomFB[rand.Intn(len(randomFB))]); err != nil {
				b.Fatal(err)
			}
		}
	})
	if err := db.Close(); err != nil {
		b.Fatal("failed to close database:", err)
	}
}
