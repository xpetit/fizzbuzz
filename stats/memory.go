package stats

import (
	"sync"

	"github.com/xpetit/fizzbuzz/v5"
)

type memory struct {
	m  map[fizzbuzz.Config]int
	mu sync.RWMutex
}

// Memory holds a protected (thread safe) hit count.
func Memory() *memory {
	return &memory{
		m: map[fizzbuzz.Config]int{},
	}
}

func (s *memory) Increment(cfg fizzbuzz.Config) error {
	s.mu.Lock()
	s.m[cfg]++
	s.mu.Unlock()
	return nil
}

func smaller(a, b fizzbuzz.Config) bool {
	if a.Limit != b.Limit {
		return a.Limit < b.Limit
	}
	if a.Int1 != b.Int1 {
		return a.Int1 < b.Int1
	}
	if a.Int2 != b.Int2 {
		return a.Int2 < b.Int2
	}
	if a.Str1 != b.Str1 {
		return a.Str1 < b.Str1
	}
	return a.Str2 < b.Str2
}

func (s *memory) MostFrequent() (count int, cfg fizzbuzz.Config, err error) {
	s.mu.RLock()
	for config, c := range s.m {
		// the configs with same count are differentiated because the "iteration order over maps is not specified" (Go spec)
		if c > count || c == count && smaller(config, cfg) {
			count = c
			cfg = config
		}
	}
	s.mu.RUnlock()
	return
}
