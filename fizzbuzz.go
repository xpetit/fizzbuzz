package fizzbuzz

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

// Config contains the Fizz buzz parameters
type Config struct {
	Limit int    `json:"limit"` // Limit is the last number of the Fizz buzz suite (1 being the first)
	Int1  int    `json:"int1"`  // Int1 is the first divisor
	Int2  int    `json:"int2"`  // Int2 is the second divisor
	Str1  string `json:"str1"`  // Str1 is the string that replaces the number when it is divisible by Int1
	Str2  string `json:"str2"`  // Str2 is the string that replaces the number when it is divisible by Int2
}

// ErrInvalidInput is returned by WriteTo when attempting to write an invalid config (negative or zero Int1/Int2).
var ErrInvalidInput = errors.New("invalid input")

// LessThan compares each exported field to determine which config is the "smallest".
func (a *Config) LessThan(b *Config) bool {
	if a.Limit != b.Limit {
		return a.Limit < b.Limit
	} else if a.Int1 != b.Int1 {
		return a.Int1 < b.Int1
	} else if a.Int2 != b.Int2 {
		return a.Int2 < b.Int2
	} else if a.Str1 != b.Str1 {
		return a.Str1 < b.Str1
	}
	return a.Str2 < b.Str2
}

// Default returns a default configuration that gives all possible types of Fizz buzz values.
func Default() *Config {
	return &Config{
		Limit: 10,
		Int1:  2,
		Int2:  3,
		Str1:  "fizz",
		Str2:  "buzz",
	}
}

// Ensure function implements interface.
var _ io.WriterTo = (*Config)(nil)

// WriteTo writes a list of Fizz buzz values as a JSON array of strings, followed by a newline character.
//
// Attempting to write a Fizz buzz with negative or zero divisors causes WriteTo to return an ErrInvalidInput.
// Any other errors reported may be due to w.WriteString or w.Write.
func (c *Config) WriteTo(w io.Writer) (n int64, err error) {
	// Check the config validity
	if c.Int1 < 1 {
		return 0, fmt.Errorf("%w: Int1 must be strictly positive", ErrInvalidInput)
	} else if c.Int2 < 1 {
		return 0, fmt.Errorf("%w: Int2 must be strictly positive", ErrInvalidInput)
	}

	// write accumulates the n bytes and returns false if the writing failed
	write := func(b []byte) bool {
		var nn int
		nn, err = w.Write(b)
		n += int64(nn)
		return err == nil
	}

	if c.Limit < 1 {
		// Fizz buzz starts with 1, so return an empty array
		write([]byte("[]\n"))
		return
	}

	// Open the JSON array
	if !write([]byte{'['}) {
		return
	}

	// Marshal JSON strings, it is safe to ignore the error because a string cannot cause one
	s1, _ := json.Marshal(c.Str1)
	s2, _ := json.Marshal(c.Str2)
	s12, _ := json.Marshal(c.Str1 + c.Str2)

	// buf is used to accumulate the bytes for a Fizz buzz JSON string
	var buf []byte

	// intBuf is a buffer big enough to hold math.MaxInt (9223372036854775807), used to format an int
	intBuf := make([]byte, 0, 19)

	// Iterate over all Fizz buzz values and write them one by one
	for i := 1; i <= c.Limit; i++ {
		// Truncate the slice while keeping the underlying storage intact to avoid unnecessary memory allocations
		buf = buf[:0]

		if i%c.Int1 == 0 {
			if i%c.Int2 == 0 {
				// i is divisible by both Int1 and Int2, append Str1+Str2 JSON string
				buf = append(buf, s12...)
			} else {
				// i is only divisible by Int1, append Str1 JSON string
				buf = append(buf, s1...)
			}
		} else if i%c.Int2 == 0 {
			// i is only divisible by Int2, append Str2 JSON string
			buf = append(buf, s2...)
		} else {
			// i is not divisible by either Int1 or Int2, append the current number i as a JSON string
			buf = append(buf, '"')
			buf = append(buf, strconv.AppendInt(intBuf, int64(i), 10)...)
			buf = append(buf, '"')
		}

		// If we haven't reached the last Fizz buzz value, add a comma (the JSON array separator)
		if i < c.Limit {
			buf = append(buf, ',')
		}

		// Finally, write the buffer
		if !write(buf) {
			return
		}

		// This prevents overflow
		if i == math.MaxInt {
			break
		}
	}
	// Close JSON array and add a newline to be consistent with (*json.Encoder).Encode
	write([]byte("]\n"))
	return
}
