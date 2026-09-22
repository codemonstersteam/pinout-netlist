# Multi-stage build SUT (template-go-api). Контекст — корень репо (PRE пуст)
# или корень воркспейса (PRE=pinout-netlist/, полигон).
# CGO=0 — чистый Go placeholder; если добавишь mattn/go-sqlite3, включи CGO.
ARG PRE=
FROM golang:1.26-alpine AS build
ARG PRE
WORKDIR /src
COPY ${PRE}go.mod ./
COPY ${PRE}go.sum* ./
RUN go mod download
COPY ${PRE} . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags '-s -w' -o /out/app ./cmd/app

FROM alpine:3.21
ARG PRE
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/app /app/app
COPY ${PRE}config.yaml /app/config.yaml
EXPOSE 8080
HEALTHCHECK --interval=2s --timeout=2s --start-period=5s --retries=10 \
    CMD wget -qO- http://localhost:8080/health || exit 1
USER app:app
ENTRYPOINT ["/app/app"]
