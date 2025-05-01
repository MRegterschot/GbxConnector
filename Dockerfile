FROM golang:1.24-alpine AS builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
EXPOSE 6980

CMD ["./server"]
