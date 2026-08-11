FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o primary .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/primary .

EXPOSE 8081

CMD ["./primary"]