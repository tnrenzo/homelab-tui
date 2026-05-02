FROM golang:1.25.8 AS build

WORKDIR /src

COPY go.mod ./
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/frontend ./frontend

FROM alpine:3.20 AS alpine-runtime

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/frontend /app/frontend

ENTRYPOINT ["/app/frontend"]

FROM debian:bookworm-slim AS debian-runtime

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /out/frontend /app/frontend

ENTRYPOINT ["/app/frontend"]
