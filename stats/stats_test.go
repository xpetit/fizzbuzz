package stats_test

import (
	"context"
	"testing"

	"github.com/xpetit/fizzbuzz/v5"
	"github.com/xpetit/fizzbuzz/v5/stats"
)

func testStatsService(t *testing.T, s stats.Service) {
	assertMostFrequentEqual := func(count int, cfg *fizzbuzz.Config, err error) {
		count2, cfg2, err2 := s.MostFrequent()

		differentCfg := (cfg != nil && cfg2 != nil && *cfg != *cfg2) ||
			(cfg == nil && cfg2 != nil) ||
			(cfg != nil && cfg2 == nil)

		if count != count2 || differentCfg || err != err2 {
			t.Errorf("MostFrequent() gave %d %v %v, want %d %v %v", count2, cfg2, err2, count, cfg, err)
		}
	}
	mustIncrement := func(cfg *fizzbuzz.Config) {
		if err := s.Increment(cfg); err != nil {
			t.Error("Increment() returned an error:", err)
		}
	}

	small := &fizzbuzz.Config{}
	big := &fizzbuzz.Config{Str2: "a"}

	assertMostFrequentEqual(0, nil, nil)

	mustIncrement(big)
	assertMostFrequentEqual(1, big, nil)

	mustIncrement(small)
	assertMostFrequentEqual(1, small, nil)

	mustIncrement(small)
	assertMostFrequentEqual(2, small, nil)
}

func TestMemory(t *testing.T) {
	testStatsService(t, &stats.Memory{})
}

func TestDB(t *testing.T) {
	db, err := stats.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	testStatsService(t, db)
	if err := db.Close(); err != nil {
		t.Fatal("failed to close database:", err)
	}
}
