package stats_test

import (
	"context"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/xpetit/fizzbuzz/v5"
	"github.com/xpetit/fizzbuzz/v5/stats"
)

func testStatsService(t *testing.T, s stats.Service) {
	mustIncrement := func(cfg fizzbuzz.Config) {
		if err := s.Increment(cfg); err != nil {
			t.Fatalf("Increment(%#v) returned an error: %v", cfg, err)
		}
	}
	assertMostFrequentEqual := func(t *testing.T, count int, cfg fizzbuzz.Config) {
		t.Helper()
		count2, cfg2, err := s.MostFrequent()
		if err != nil {
			t.Fatal("MostFrequent() returned an error:", err)
		}
		if count != count2 || cfg != cfg2 {
			t.Errorf("MostFrequent() gave %d %v, nil want %d %v, nil", count2, cfg2, count, cfg)
		}
	}

	var small fizzbuzz.Config
	big := fizzbuzz.Config{Str2: "a"}

	mustIncrement(big)
	assertMostFrequentEqual(t, 1, big)

	mustIncrement(small)
	assertMostFrequentEqual(t, 1, small)

	mustIncrement(small)
	assertMostFrequentEqual(t, 2, small)
}

func TestMemory(t *testing.T) {
	testStatsService(t, stats.Memory())
}

func TestDB(t *testing.T) {
	db, err := stats.OpenDB(context.Background(), ":memory:")
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	testStatsService(t, db)
	if err := db.Close(); err != nil {
		t.Fatal("failed to close database:", err)
	}
}

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

	db, err := stats.OpenDB(context.Background(), filepath.Join(b.TempDir(), "data.db"))
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
