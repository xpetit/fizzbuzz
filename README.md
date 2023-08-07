# FizzBuzz

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

## Usage

Requirements:

- [Go 1.20 or newer](https://golang.org/dl/)

Use this command to directly update and run the service:

```
go run github.com/xpetit/fizzbuzz/v5/cmd/fizzbuzzd@latest
```

To stop it, type <kbd>CTRL</kbd>-<kbd>C</kbd>.

If you don't trust this program, you can use Docker. Clone this repository and run the following commands inside:

```
docker build --tag github.com/xpetit/fizzbuzz/v5 .
docker run --rm --publish 8080:8080 github.com/xpetit/fizzbuzz/v5
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
2. A library that streams JSON strings and one that generates them all at once are a very different story. One cannot be the generalized/abstract version of another.

![relevant XKCD](https://imgs.xkcd.com/comics/the_general_problem.png)

### Packages

The top-down list of dependencies is as follows:

- `github.com/xpetit/fizzbuzz/v5/cmd/fizzbuzzd`: The main program, running the HTTP server.
- `github.com/xpetit/fizzbuzz/v5/handlers`: The HTTP handlers.
- `github.com/xpetit/fizzbuzz/v5/stats`: The statistics services.
- `github.com/xpetit/fizzbuzz/v5`: The Fizz buzz writer `WriteTo`.

### Performance

Benchmarks performed on an AWS EC2 instance (c6i.4xlarge) with the following specs:

|     |                                        |
| --- | -------------------------------------- |
| CPU | Intel Xeon Platinum 8375C CPU @2.90GHz |
| RAM | 32 GB ECC 2666 MHz                     |
| Go  | Version 1.20.5                         |

The programs were compiled with `GOAMD64=v4` and ran with `GOGC=1000` environment variables.

`WriteTo` limits memory allocation, here are the results from 0 to 10 million values of Fizz buzz:

```
BenchmarkWriteTo/[limit:0e+00]-16   271018197       4.431 ns/op   677.05 MB/s    0 B/op   0 allocs/op
BenchmarkWriteTo/[limit:1e+00]-16    11806252       99.79 ns/op    60.13 MB/s   32 B/op   5 allocs/op
BenchmarkWriteTo/[limit:1e+01]-16     5970025       200.2 ns/op   334.67 MB/s   48 B/op   6 allocs/op
BenchmarkWriteTo/[limit:1e+02]-16     1000000        1016 ns/op   685.71 MB/s   48 B/op   6 allocs/op
BenchmarkWriteTo/[limit:1e+03]-16      114363       10508 ns/op   694.45 MB/s   48 B/op   6 allocs/op
BenchmarkWriteTo/[limit:1e+04]-16       10000      106170 ns/op   718.63 MB/s   48 B/op   6 allocs/op
BenchmarkWriteTo/[limit:1e+05]-16        1074     1113399 ns/op   715.19 MB/s   48 B/op   6 allocs/op
BenchmarkWriteTo/[limit:1e+06]-16         100    11305590 ns/op   733.82 MB/s   48 B/op   6 allocs/op
BenchmarkWriteTo/[limit:1e+07]-16           9   117342031 ns/op   735.43 MB/s   49 B/op   6 allocs/op
```

The service stops writing values as soon as the API consumer no longer requests them.

[`wrk`](https://github.com/wg/wrk) reports 447k requests/second with a random limit between 1 and 100 (using `-db off` command-line argument):

```
Running 10s test @ http://127.0.0.1:8080
  5 threads and 600 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.45ms    1.85ms  47.51ms   90.65%
    Req/Sec    90.68k     7.72k  111.89k    68.94%
  4511934 requests in 10.10s, 1.86GB read
Requests/sec: 446918.14
Transfer/sec:    188.49MB
```

Leaving the database enabled:

```
Running 10s test @ http://127.0.0.1:8080
  4 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    39.17ms   70.20ms 770.02ms   88.86%
    Req/Sec     7.81k   636.24    10.35k    76.25%
  310796 requests in 10.02s, 131.31MB read
Requests/sec:  31002.22
Transfer/sec:     13.10MB
```

And with `-db :memory:` command-line argument ([SQLite in-memory DB](https://www.sqlite.org/inmemorydb.html)):

```
Running 10s test @ http://127.0.0.1:8080
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   242.48us  189.82us   2.21ms   86.00%
    Req/Sec    22.80k   635.84    24.37k    65.84%
  458245 requests in 10.10s, 193.35MB read
Requests/sec:  45371.08
Transfer/sec:     19.14MB
```
