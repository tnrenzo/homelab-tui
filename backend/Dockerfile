FROM golang:1.25.8 AS build

WORKDIR /src

COPY go.mod ./
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/backend ./backend

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/backend /app/backend

EXPOSE 8080

CMD ["/app/backend"]
