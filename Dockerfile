FROM golang:1.22-alpine AS builder

RUN go install github.com/a-h/templ/cmd/templ@v0.3.857

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN templ generate && \
    go build -o /app/bin/webmail ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/webmail ./webmail
COPY --from=builder /app/static      ./static

VOLUME ["/app/config.json"]

EXPOSE 8080

ENTRYPOINT ["./webmail", "/app/config.json"]