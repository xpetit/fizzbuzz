# syntax=docker/dockerfile:1

# For more information, please visit: https://docs.docker.com/language/golang/build-images

# Leverage multi-stage build to reduce the final Docker image size
FROM golang:1.17-alpine as builder

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
RUN adduser -D user
USER user:user

# Copy binary
ENTRYPOINT ["/app/fizzbuzzd"]
COPY --from=builder /app/fizzbuzzd /app/fizzbuzzd

# The HTTP listening port is both configurable at build-time (image) and runtime (container):
#   docker build --build-arg PORT=8081 --tag github.com/xpetit/fizzbuzz/v5
# or:
#   docker run --rm --env PORT=8081 --publish 9000:8081 github.com/xpetit/fizzbuzz/v5
ARG PORT=8080
ENV PORT ${PORT}
EXPOSE ${PORT}
