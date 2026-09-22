# Multi-stage build SUT (template-go-api). Контекст — корень репо.
# CGO=0 — чистый Go placeholder; если добавишь mattn/go-sqlite3, включи CGO.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY go.sum* ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags '-s -w' -o /out/app ./cmd/app

FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/app /app/app
COPY config.yaml /app/config.yaml
EXPOSE 8080
HEALTHCHECK --interval=2s --timeout=2s --start-period=5s --retries=10 \
    CMD wget -qO- http://localhost:8080/health || exit 1
USER app:app
ENTRYPOINT ["/app/app"]
