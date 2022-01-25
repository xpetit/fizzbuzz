package main

import (
	"flag"
	"os"

	"github.com/xpetit/fizzbuzz/v4"
)

func main() {
	c := fizzbuzz.Default()
	flag.IntVar(&c.Limit, "limit", c.Limit, "Limit is the last number of the Fizz buzz suite (1 being the first)")
	flag.IntVar(&c.Int1, "int1", c.Int1, "Int1 is the first divisor")
	flag.IntVar(&c.Int2, "int2", c.Int2, "Int2 is the second divisor")
	flag.StringVar(&c.Str1, "str1", c.Str1, "Str1 is the string that replaces the number when it is divisible by Int1")
	flag.StringVar(&c.Str2, "str2", c.Str2, "Str2 is the string that replaces the number when it is divisible by Int2")
	flag.Parse()
	if _, err := c.WriteTo(os.Stdout); err != nil {
		panic(err)
	}
}
