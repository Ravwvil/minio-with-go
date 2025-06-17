FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download


RUN go build -o app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app/app /app

EXPOSE 8080

CMD ["/app"]