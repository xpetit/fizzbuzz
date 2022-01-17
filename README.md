# FizzBuzz

FizzBuzz is a Golang HTTP server exposing a RESTful web API.

It has two endpoints:

- `/api/v1/fizzbuzz`
  - Accepts five optional query parameters : three integers `int1`, `int2` and `limit`, and two strings `str1` and `str2`.<br>
    The default values are: `limit=10`, `int1=2`, `int2=3`, `str1=fizz`, `str2=buzz`.
  - Returns a list of strings with numbers from 1 to `limit`, where: all multiples of `int1` are replaced by `str1`, all multiples of `int2` are replaced by `str2`, all multiples of `int1` and `int2` are replaced by `str1str2`.
- `/api/v1/fizzbuzz/stats`
  - Accept no parameters
  - Return the parameters corresponding to the most used request, as well as the number of hits for this request

The server is:

- Ready for production
- Easy to maintain by other developers

## Usage

Requirements:

- [Go 1.17 or newer](https://golang.org/dl/)

Use this command to directly update and run the service:

```
GOPROXY=direct go run github.com/xpetit/fizzbuzz/cmd/fizzbuzz@latest
```

To stop it, type <kbd>CTRL</kbd>-<kbd>C</kbd>.

If you don't trust this program, you can use Docker, clone this repository and run the following commands inside:

```
docker build --tag github.com/xpetit/fizzbuzz .
docker run --rm --publish 8080:8080 github.com/xpetit/fizzbuzz
```

When the service is running, you can query with it with `curl`:

```console
user@host$ curl localhost:8080/api/v1/fizzbuzz/stats
{"most_frequent":null,"count":0}
user@host$ curl localhost:8080/api/v1/fizzbuzz
["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
user@host$ curl localhost:8080/api/v1/fizzbuzz/stats
{"most_frequent":{"limit":10,"int1":2,"int2":3,"str1":"fizz","str2":"buzz"},"count":1}
```

## Design

In writing this library, several considerations were taken into account:

1. Should it provide a generalized, extensible implementation that supports any number of "int" and "str"?
2. Should it provide a FizzBuzz function that not only writes values but returns values?
   1. Should the returned values be a slice of JSON string or a more compact (binary) representation for later reuse?

This was discarded for the following reasons:

1. If requirements change, it is easy to adapt the Fizz buzz algorithm because it is simple.
2. As described in the [Performance](#performance) section, a library that streams JSON strings and one that generates them all at once are a very different story. One cannot be the generalized/abstract version of another.

![relevant XKCD](https://imgs.xkcd.com/comics/the_general_problem.png)

### Performance

For a Fizz buzz with a limit of one million, the naive approach (`WriteInto2`) takes about 5 times longer and uses 110 MB of memory, while the optimized implementation (`WriteInto`) does not allocate memory:

```
BenchmarkWriteInto/big-12    58    20081942 ns/op         142 B/op        9 allocs/op
BenchmarkWriteInto2/big-12   12   100073916 ns/op   109733862 B/op   500030 allocs/op
```

However the biggest problem with the naive implementation is that it first generates the JSON array and then writes it, even if the writer has been closed before.
This means that a buggy program looping through this API can create unnecessary work and resource exhaustion. The same is true for an attacker.
The optimized implementation stops writing Fizz buzz values as soon as the API consumer no longer requests them.

An average of 190 MB/s actual throughput and 400 MB/s theoretical maximum throughput was measured in the following dedicated benchmark environment:

|     |                                                  |
| --- | ------------------------------------------------ |
| CPU | Intel Xeon E-2236 CPU @3.40GHz                   |
| RAM | 32 GB ECC 2666 MHz                               |
| OS  | Debian Stable (`scaling_governor = performance`) |
| Go  | Version 1.17.5                                   |

The command used to measure HTTP throughput:

```
curl -o /dev/null localhost:8080/api/v1/fizzbuzz?limit=10000000000
```

In addition, [`wrk`](https://github.com/wg/wrk) reports 221734 requests/second with a random limit between 1 and 100:

```
Running 10s test @ http://localhost:8080/api/v1/fizzbuzz?limit=param_value
  4 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.14ms    2.52ms  29.34ms   85.42%
    Req/Sec    55.99k     5.89k   71.20k    68.25%
  2238244 requests in 10.09s, 0.99GB read
Requests/sec: 221734.27
Transfer/sec:    100.36MB
```
