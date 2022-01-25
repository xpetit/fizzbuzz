# FizzBuzz

![coverage](./coverage.svg)

FizzBuzz is a Golang HTTP server exposing a RESTful web API that provides a [Fizz buzz](https://en.wikipedia.org/wiki/Fizz_buzz) service.

It has two endpoints:

- `/api/v2/fizzbuzz`
  - Accepts five optional query parameters : three integers `int1`, `int2` and `limit`, and two strings `str1` and `str2`.<br>
    The default values are: `limit=10`, `int1=2`, `int2=3`, `str1=fizz`, `str2=buzz`.
  - Returns a list of strings with numbers from 1 to `limit`, where: all multiples of `int1` are replaced by `str1`, all multiples of `int2` are replaced by `str2`, all multiples of `int1` and `int2` are replaced by `str1str2`.
- `/api/v2/fizzbuzz/stats`
  - Accept no parameters
  - Return the parameters corresponding to the most used request, as well as the number of hits for this request

The server is:

- Ready for production
- Easy to maintain by other developers
- Not using external libraries (as requested)

## Usage

Requirements:

- [Go 1.17 or newer](https://golang.org/dl/)

Use this command to directly update and run the service:

```
go run github.com/xpetit/fizzbuzz/v4/cmd/fizzbuzzd@latest
```

To stop it, type <kbd>CTRL</kbd>-<kbd>C</kbd>.

If you don't trust this program, you can use Docker. Clone this repository and run the following commands inside:

```
docker build --tag github.com/xpetit/fizzbuzz/v4 .
docker run --rm --publish 8080:8080 github.com/xpetit/fizzbuzz/v4
```

When the service is running, you can query with it with `curl`:

```
curl localhost:8080/api/v2/fizzbuzz/stats
```

> <!-- prettier-ignore -->
> ```json
> {"most_frequent":{"count":0}}
> ```

Using the defaults:

```
curl localhost:8080/api/v2/fizzbuzz
```

> <!-- prettier-ignore -->
> ```json
> ["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
> ```

The `config` object is now populated:

```
curl localhost:8080/api/v2/fizzbuzz/stats
```

> <!-- prettier-ignore -->
> ```json
> {"most_frequent":{"count":1,"config":{"limit":10,"int1":2,"int2":3,"str1":"fizz","str2":"buzz"}}}
> ```

Custom request:

```
curl localhost:8080/api/v2/fizzbuzz -Gdlimit=1 -dint1=1 -dint2=1 -dstr1=buzz -dstr2=lightyear
```

> ```json
> ["buzzlightyear"]
> ```

## Design

In writing this library, several considerations were taken into account:

1. Should it provide a generalized, extensible implementation that supports any number of "int" and "str"?
2. Should it provide a FizzBuzz function that not only writes values but returns values?
   1. Should the returned values be a slice of JSON string or a more compact (binary) representation for later reuse?

This was discarded for the following reasons:

1. If requirements change, it is easy to adapt the Fizz buzz algorithm because it is simple.
2. As described in the [Performance](#performance) section, a library that streams JSON strings and one that generates them all at once are a very different story. One cannot be the generalized/abstract version of another.

![relevant XKCD](https://imgs.xkcd.com/comics/the_general_problem.png)

### Packages

The top-down list of dependencies is as follows:

- `github.com/xpetit/fizzbuzz/v4/cmd/fizzbuzzd`: The main program, running the HTTP server.
- `github.com/xpetit/fizzbuzz/v4/handlers`: The HTTP handlers.
- `github.com/xpetit/fizzbuzz/v4`: The Fizz buzz writer `WriteTo`.

### Performance

`WriteTo` limits memory allocation, here are the results from 0 to 10 million values of Fizz buzz:

```
BenchmarkWriteTo/[limit:0e+00]-12   100000000          11 ns/op   268.82 MB/s     3 B/op   1 allocs/op
BenchmarkWriteTo/[limit:1e+00]-12     3801452         314 ns/op    19.12 MB/s    80 B/op   8 allocs/op
BenchmarkWriteTo/[limit:1e+01]-12     2356630         509 ns/op   131.71 MB/s    96 B/op   9 allocs/op
BenchmarkWriteTo/[limit:1e+02]-12      564086        2141 ns/op   325.50 MB/s    96 B/op   9 allocs/op
BenchmarkWriteTo/[limit:1e+03]-12       55982       19697 ns/op   370.47 MB/s    96 B/op   9 allocs/op
BenchmarkWriteTo/[limit:1e+04]-12        6004      195040 ns/op   391.19 MB/s    96 B/op   9 allocs/op
BenchmarkWriteTo/[limit:1e+05]-12         589     1999506 ns/op   398.25 MB/s    98 B/op   9 allocs/op
BenchmarkWriteTo/[limit:1e+06]-12          57    20156702 ns/op   411.59 MB/s   123 B/op   9 allocs/op
BenchmarkWriteTo/[limit:1e+07]-12           5   205326151 ns/op   420.29 MB/s   408 B/op   9 allocs/op
```

The service stops writing values as soon as the API consumer no longer requests them.

An average of 190 MB/s actual throughput was measured in the following dedicated benchmark environment:

|     |                                                  |
| --- | ------------------------------------------------ |
| CPU | Intel Xeon E-2236 CPU @3.40GHz                   |
| RAM | 32 GB ECC 2666 MHz                               |
| OS  | Debian Stable (`scaling_governor = performance`) |
| Go  | Version 1.17.5                                   |

The command used to measure HTTP throughput:

```
curl -o /dev/null localhost:8080/api/v2/fizzbuzz?limit=10000000000
```

In addition, [`wrk`](https://github.com/wg/wrk) reports 250151 requests/second with a random limit between 1 and 100:

```
Running 10s test @ http://localhost:8080/api/v2/fizzbuzz?limit=param_value
  4 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.02ms    3.25ms  44.49ms   89.50%
    Req/Sec    63.27k    14.84k   86.27k    57.83%
  2517462 requests in 10.06s, 1.11GB read
Requests/sec: 250150.61
Transfer/sec:    113.18MB
```
