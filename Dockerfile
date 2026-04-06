# syntax=docker/dockerfile:1

FROM golang:1.20-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS=linux
ARG TARGETARCH
RUN set -eux; \
    if [ -n "${TARGETARCH:-}" ]; then export GOARCH="${TARGETARCH}"; fi; \
    CGO_ENABLED=0 GOOS="${TARGETOS}" go build -trimpath -ldflags="-s -w" -o /out/aa-crystals-calc-bot ./cmd/bot

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && \
    adduser -S -G app app

WORKDIR /app

COPY --from=build /out/aa-crystals-calc-bot /app/aa-crystals-calc-bot

ENV USD_RUB_FALLBACK=80
ENV MAX_SHK=100000
ENV METRICS_ADDR=0.0.0.0:8080

EXPOSE 8080

USER app

ENTRYPOINT ["/app/aa-crystals-calc-bot"]
