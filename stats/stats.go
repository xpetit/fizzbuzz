package stats

import "github.com/xpetit/fizzbuzz/v5"

type Service interface {
	Increment(cfg fizzbuzz.Config) error
	MostFrequent() (count int, cfg fizzbuzz.Config, err error)
}

var (
	_ Service = (*memory)(nil)
	_ Service = (*db)(nil)
)
