package fizzbuzz

import (
	"encoding/json"
	"io"
	"strconv"
)

type Config struct {
	Limit int    `json:"limit"` // Limit is the last number (1 being the first)
	X     int    `json:"int1"`  // X is the first divider
	Y     int    `json:"int2"`  // Y is the second divider
	A     string `json:"str1"`  // A is the string to use when the number is divisible by X
	B     string `json:"str2"`  // B is the string to use when the number is divisible by Y
}

func (c *Config) toString(i int) string {
	if i%c.X == 0 {
		if i%c.Y == 0 {
			// i is divisible by both x and y
			return c.A + c.B
		}
		// i is only divisible by x
		return c.A
	}
	if i%c.Y == 0 {
		// i is only divisible by y
		return c.B
	}
	// i is not divisible by either x or y
	return strconv.Itoa(i)
}

func (c *Config) ToSlice() (ss []string) {
	for i := 1; i <= c.Limit; i++ {
		ss = append(ss, c.toString(i))
	}
	return
}

// Ensure type implements interface.
var _ io.WriterTo = (*Config)(nil)

func (c *Config) WriteTo(w io.Writer) (int64, error) {
	n, err := c.writeTo(w)
	return int64(n), err
}

func (c *Config) writeTo(w io.Writer) (n int, err error) {
	// Check the config validity
	if c.Limit < 1 {
		// Fizz buzz starts with 1, so this is empty
		return io.WriteString(w, "[]\n")
	} else if c.X < 1 || c.Y < 1 {
		panic("X and Y must be strictly positive")
	}

	// Open the JSON array
	if n, err = io.WriteString(w, "["); err != nil {
		return
	}

	// Marshal JSON strings, it is safe to ignore the error because a string cannot cause one
	ab, _ := json.Marshal(c.A + c.B)
	a, _ := json.Marshal(c.A)
	b, _ := json.Marshal(c.B)

	// Iterate over all Fizz buzz values and write them
	var buf []byte // buf is used to accumulate the bytes and write them all at once
	for i := 1; i <= c.Limit; i++ {
		// Truncate the slice while keeping the underlying storage intact to avoid unnecessary memory allocations
		buf = buf[:0]

		if i%c.X == 0 {
			if i%c.Y == 0 {
				// i is divisible by both X and Y, append A+B JSON string
				buf = append(buf, ab...)
			} else {
				// i is only divisible by X, append A JSON string
				buf = append(buf, a...)
			}
		} else if i%c.Y == 0 {
			// i is only divisible by Y, append B JSON string
			buf = append(buf, b...)
		} else {
			// i is not divisible by either X or Y, append the current number i as a JSON string
			buf = append(buf, '"')
			buf = append(buf, strconv.Itoa(i)...)
			buf = append(buf, '"')
		}

		// If we haven't reached the last Fizz buzz value, add a comma (the JSON array separator)
		if i != c.Limit {
			buf = append(buf, ',')
		}

		// Finally, write the buffer and update the number of bytes written
		nn, err := w.Write(buf)
		n += int(nn)
		if err != nil {
			return n, err
		}
	}
	// Close JSON array and add a newline to be consistent with (*json.Encoder).Encode
	nn, err := io.WriteString(w, "]\n")
	return n + nn, err
}
