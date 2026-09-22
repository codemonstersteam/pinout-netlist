# Стейджер бинаря валидатора из братского репо воркспейса (ARG REPO).
# Собирается контекстом КОРНЯ воркспейса: COPY ${REPO}/... (см. docker-compose.polygon.yml).
ARG REPO
FROM golang:1.26-alpine AS build
ARG REPO
WORKDIR /src
COPY ${REPO}/go.mod ${REPO}/go.sum* ./
RUN go mod download
COPY ${REPO}/ .
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o /out/tool ./cmd/app

FROM alpine:3.21
ARG NAME=tool
ENV NAME=${NAME}
COPY --from=build /out/tool /out/tool
# Имя бинаря в /tools задаёт переменная NAME (compose передаёт по репо).
ENTRYPOINT ["sh", "-c", "cp /out/tool /tools/${NAME}"]
