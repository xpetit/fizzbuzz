# FizzBuzz

FizzBuzz is a Golang HTTP server exposing a RESTful web API.

It exposes two endpoints:

- `/api/v1/fizzbuzz`
  - Accepts five optional query parameters : three integers `int1`, `int2` and `limit`, and two strings `str1` and `str2`.<br>
    The default values are: `limit=10`, `int1=2`, `int2=3`, `str1=fizz`, `str2=buzz`.
  - Returns a list of strings with numbers from 1 to `limit`, where: all multiples of `int1` are replaced by `str1`, all multiples of `int2` are replaced by `str2`, all multiples of `int1` and `int2` are replaced by `str1str2`.
- `/api/v1/statistics`
  - Accept no parameters
  - Return the parameters corresponding to the most used request, as well as the number of hits for this request

The server is:

- Ready for production
- Easy to maintain by other developers

## Build from source

Requirements:

- [Go 1.17 or newer](https://golang.org/dl/)

```
go build ./cmd/fizzbuzz
```

## Build Docker image

```
docker build --tag github.com/xpetit/fizzbuzz .
```

## Start service

If you have followed [Build from source](#Build-from-source):

```
./fizzbuzz
```

Or if you have followed [Build Docker image](#Build-Docker-image):

```
docker run --rm --publish 8080:8080 github.com/xpetit/fizzbuzz
```

## Stop service

To stop the service, type <kbd>CTRL</kbd>-<kbd>C</kbd>.

## Usage

When the service is running, you can query with it with `curl`:

```console
user@host$ curl localhost:8080/api/v1/statistics
{"most_frequent":null,"count":0}
user@host$ curl localhost:8080/api/v1/fizzbuzz
["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
user@host$ curl localhost:8080/api/v1/statistics
{"most_frequent":{"limit":10,"int1":2,"int2":3,"str1":"fizz","str2":"buzz"},"count":1}
```

## Implementation
