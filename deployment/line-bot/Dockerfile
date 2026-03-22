FROM --platform=linux/amd64 golang:1.26.0-alpine3.23 AS builder
# FROM golang:1.26.0-alpine3.23 AS builder

COPY ./ /service/app

WORKDIR /service/app

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/line-bot/main.go
# RUN go build -o app cmd/line-bot/main.go

FROM --platform=linux/amd64 alpine:3.23 AS runner
# FROM alpine:3.23 AS runner

COPY --from=builder /service/app/app /app

EXPOSE 8080

ENTRYPOINT ["./app"]