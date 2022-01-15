# FizzBuzz

FizzBuzz is a Golang HTTP server exposing a RESTful web API.

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
user@host$ curl localhost:8080/api/v1/fizzbuzz
["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]
user@host$ curl localhost:8080/api/v1/statistics
{"most_frequent":{"limit":10,"int1":3,"int2":5,"str1":"fizz","str2":"buzz"}}
```

## Implementation
