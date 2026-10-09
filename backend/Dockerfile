FROM golang:1.27.2-alpine3.23 AS build

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY backend ./backend
COPY database ./database

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=1 GOMEMLIMIT=512MiB CGO_ENABLED=1 GOOS=linux \
    go build -p=1 -trimpath -ldflags="-s -w" -o /out/api ./backend/cmd/api

FROM alpine:3.23

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 10001 dulio && \
    adduser -S -D -H -u 10001 -G dulio dulio && \
    mkdir -p /app /data/tls && \
    chown -R dulio:dulio /app /data

COPY --from=build --chown=dulio:dulio /out/api /app/api

USER dulio
WORKDIR /app

ENV DATABASE_PATH=/data/dulio.db

EXPOSE 443

ENTRYPOINT ["/app/api"]
