package fizzbuzz

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

type Config struct {
	Limit int    `json:"limit"` // Limit is the last number (1 being the first)
	Int1  int    `json:"int1"`  // Int1 is the first divider
	Int2  int    `json:"int2"`  // Int2 is the second divider
	Str1  string `json:"str1"`  // Str1 is the string to use when the number is divisible by Int1
	Str2  string `json:"str2"`  // Str2 is the string to use when the number is divisible by Int2
}

var ErrInvalidInput = errors.New("invalid input")

func (c *Config) toString(i int) string {
	if i%c.Int1 == 0 {
		if i%c.Int2 == 0 {
			// i is divisible by both Int1 and Int2
			return c.Str1 + c.Str2
		}
		// i is only divisible by Int1
		return c.Str1
	}
	if i%c.Int2 == 0 {
		// i is only divisible by Int2
		return c.Str2
	}
	// i is not divisible by either Int1 or Int2
	return strconv.Itoa(i)
}

func (c *Config) WriteWith2(w io.Writer) error {
	// Check the config validity
	if c.Int1 < 1 {
		return fmt.Errorf("%w: Int1 must be strictly positive", ErrInvalidInput)
	} else if c.Int2 < 1 {
		return fmt.Errorf("%w: Int2 must be strictly positive", ErrInvalidInput)
	}

	var ss []string
	for i := 1; i <= c.Limit; i++ {
		ss = append(ss, c.toString(i))
	}
	return json.NewEncoder(w).Encode(ss)
}

func (c *Config) WriteWith(w io.Writer) error {
	// Check the config validity
	if c.Int1 < 1 {
		return fmt.Errorf("%w: Int1 must be strictly positive", ErrInvalidInput)
	} else if c.Int2 < 1 {
		return fmt.Errorf("%w: Int2 must be strictly positive", ErrInvalidInput)
	}

	if c.Limit < 1 {
		// Fizz buzz starts with 1, so return an empty array
		_, err := io.WriteString(w, "[]\n")
		return err
	}

	// Open the JSON array
	if _, err := io.WriteString(w, "["); err != nil {
		return err
	}

	// Marshal JSON strings, it is safe to ignore the error because a string cannot cause one
	s12, _ := json.Marshal(c.Str1 + c.Str2)
	s1, _ := json.Marshal(c.Str1)
	s2, _ := json.Marshal(c.Str2)

	// buf is used to accumulate the bytes for a Fizz buzz JSON string
	var buf []byte

	// intBuf is a buffer big enough to hold math.MaxInt (9223372036854775807), used to format an int
	intBuf := make([]byte, 0, 19)

	// Iterate over all Fizz buzz values and write them one by one
	for i := 1; i <= c.Limit && i < math.MaxInt; i++ {
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
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}
	// Close JSON array and add a newline to be consistent with (*json.Encoder).Encode
	_, err := io.WriteString(w, "]\n")
	return err
}
