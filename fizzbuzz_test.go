package fizzbuzz_test

import (
	"fmt"
	"io"
	"math"
	"os"
	"testing"

	"github.com/xpetit/fizzbuzz/v5"
)

func Example() {
	c := fizzbuzz.Default()
	c.WriteTo(os.Stdout)

	c = fizzbuzz.Config{Limit: 15, Int1: 3, Int2: 5, Str1: "a", Str2: "b"}
	c.WriteTo(os.Stdout)

	// Output:
	// ["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
	// ["1","2","a","4","b","a","7","8","a","b","11","a","13","14","ab"]
}

// BenchmarkWriteTo benchmarks WriteTo with a default config and a limit of n
func BenchmarkWriteTo(b *testing.B) {
	c := fizzbuzz.Default()
	for n := -1; n <= 7; n++ {
		c.Limit = int(math.Pow10(n))
		b.Run(fmt.Sprintf("[limit:%.e]", float64(c.Limit)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if n, err := c.WriteTo(io.Discard); err != nil {
					b.Fatal(err)
				} else {
					b.SetBytes(n)
				}
			}
		})
	}
}
