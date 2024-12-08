package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

var (
	expiryTime          time.Time
	mu                  sync.Mutex
	uniqueRequestsCount int
	redisClient         *redis.Client
	kafkaWriter         *kafka.Writer
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	kafkaBrokerAddr := os.Getenv("KAFKA_ADDR")
	kafkaTopicName := os.Getenv("KAFKA_TOPIC")

	if redisAddr == "" || kafkaBrokerAddr == "" {
		log.Fatal("REDIS_ADDR and KAFKA_ADDR environment variables must be set")
	}

	connectRedisClient(redisAddr)
	connectKafkaWriter(kafkaBrokerAddr, kafkaTopicName)
	defer redisClient.Close()
	defer kafkaWriter.Close()

	router := gin.Default()

	router.GET("/api/verve/accept", acceptHandler)

	// Periodically publish unique request count to Kafka from this instance of the service
	go func() {
		for {
			expiryTime = time.Now().Add(time.Minute)
			time.Sleep(time.Minute)
			mu.Lock()
			count := uniqueRequestsCount
			uniqueRequestsCount = 0
			mu.Unlock()

			// Publish to Kafka
			if err := sendToKafka(count, expiryTime); err != nil {
				log.Printf("Failed to send to Kafka: %v", err)
			}
		}
	}()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func connectRedisClient(redisAddr string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func connectKafkaWriter(kafkaBrokerAddr, topicName string) *kafka.Writer {
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokerAddr),
		Topic:    topicName,
		Balancer: &kafka.LeastBytes{},
	}
	return kafkaWriter
}

func acceptHandler(c *gin.Context) {
	idParam := c.Query("id")
	if idParam == "" {
		c.String(http.StatusBadRequest, "failed")
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.String(http.StatusBadRequest, "failed")
		return
	}

	// check if the id is already present in the redis
	ctx := context.Background()
	present, err := redisClient.Exists(ctx, fmt.Sprintf("%s", id)).Result()
	if err != nil {
		c.String(http.StatusInternalServerError, "failed")
		return
	}
	if present == 0 {
		mu.Lock()
		uniqueRequestsCount++
		mu.Unlock()
	}

	optionalEndpoint := c.Query("endpoint")

	// redis key set to expire at the given expiry time
	_, err = redisClient.SetNX(ctx, fmt.Sprintf("%s", id), true, time.Until(expiryTime)).Result()

	if err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, "failed")
		return
	}

	// If optional endpoint is provided, make an HTTP POST request
	if optionalEndpoint != "" {
		go sendPostRequest(optionalEndpoint, uniqueRequestsCount)
	}

	c.String(http.StatusOK, "ok")
}

func sendPostRequest(endpoint string, count int) {
	data := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"count":     count,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send POST request to %s: %v", endpoint, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("POST request to %s returned status code: %d", endpoint, resp.StatusCode)
}

func sendToKafka(count int, expiryTime time.Time) error {
	message := kafka.Message{
		// add time here
		Value: []byte(fmt.Sprintf("Minute: %s, %d", expiryTime.String(), count)),
	}
	return kafkaWriter.WriteMessages(context.Background(), message)
}
