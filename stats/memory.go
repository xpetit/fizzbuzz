package stats

import (
	"sync"

	"github.com/xpetit/fizzbuzz/v5"
)

type memory struct {
	mu sync.RWMutex
	m  map[fizzbuzz.Config]int
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

func (s *memory) MostFrequent() (count int, cfg fizzbuzz.Config, err error) {
	s.mu.RLock()
	for config, c := range s.m {
		// the configs with same count are differentiated because the "iteration order over maps is not specified" (Go spec)
		if c > count || c == count && config.LessThan(cfg) {
			count = c
			cfg = config
		}
	}
	s.mu.RUnlock()
	return
}
