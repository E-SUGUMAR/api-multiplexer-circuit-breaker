FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o secondary .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/secondary .

EXPOSE 8082

CMD ["./secondary"]