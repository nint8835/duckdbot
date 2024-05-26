FROM golang:1.22-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o duckdbot .

FROM cgr.dev/chainguard/glibc-dynamic

WORKDIR /duckdbot

COPY --from=builder /build/duckdbot .

ENTRYPOINT ["/duckdbot/duckdbot"]
