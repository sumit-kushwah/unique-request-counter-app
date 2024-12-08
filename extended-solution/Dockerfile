FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app .

ENV REDIS_ADDR=redis-server:6379
ENV KAFKA_ADDR=kafka-broker:9092
ENV KAFKA_TOPIC=UNIQUE_REQUEST_COUNT_BY_MINUTE

EXPOSE 8080

CMD ["./app"]
