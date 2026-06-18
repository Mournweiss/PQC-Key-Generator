# Stage 1: Build the Go keygen binary
FROM docker.io/library/golang:1.25.3-alpine AS build

WORKDIR /keygen

COPY go.mod ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 go build -o keygen ./cmd/keygen

# Stage 2: Runtime image using pre-built OQS-OpenSSL
FROM ghcr.io/mournweiss/oqs-openssl:latest-alpine AS app

COPY --from=build /keygen/keygen /keygen/keygen

WORKDIR /keygen

ENTRYPOINT ["./keygen"]
