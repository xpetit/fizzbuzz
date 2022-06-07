# syntax=docker/dockerfile:1

# For more information, please visit: https://docs.docker.com/language/golang/build-images

# Leverage multi-stage build to reduce the final Docker image size
FROM golang:1.18-alpine as builder

# needed for cgo github.com/mattn/go-sqlite3 dependency
RUN apk add --no-cache build-base

WORKDIR /app

# Download and cache all dependencies of the main module
COPY go.mod go.sum ./
RUN go mod download

# Build program
COPY *.go ./
COPY cmd cmd
COPY handlers handlers
COPY stats stats
# -ldflags "-s -w" reduces the binary size (-s: disable symbol table, -w: disable DWARF generation)
RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg \
	go build -ldflags "-s -w" ./cmd/fizzbuzzd


FROM alpine

# Create unprivileged user for the service
# -D: Don't assign a password
RUN adduser -u 10000 -D user
USER user:user

# Copy binary
ENTRYPOINT ["/app/fizzbuzzd"]
CMD ["--host", "0.0.0.0"]
COPY --from=builder /app/fizzbuzzd /app/fizzbuzzd

EXPOSE 8080/tcp
