package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	logger              *log.Logger
	uniqueRequestsCount int
	uniqueRequests      = make(map[int]bool)
	mu                  sync.Mutex
)

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

	optionalEndpoint := c.Query("endpoint")

	mu.Lock()
	if !uniqueRequests[id] {
		uniqueRequests[id] = true
		uniqueRequestsCount++
	}
	mu.Unlock()

	if optionalEndpoint != "" {
		go func(endpoint string) {
			resp, err := http.Get(fmt.Sprintf("%s?count=%d", endpoint, uniqueRequestsCount))
			if err != nil {
				logger.Printf("Failed to send GET request to %s: %v", endpoint, err)
				return
			}
			logger.Printf("GET request to %s returned status code: %d", endpoint, resp.StatusCode)
			resp.Body.Close()
		}(optionalEndpoint)
	}

	c.String(http.StatusOK, "ok")
}

func main() {
	// Logging Setup
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	router := gin.Default()

	router.GET("/api/verve/accept", acceptHandler)

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			mu.Lock()
			count := uniqueRequestsCount
			uniqueRequests = make(map[int]bool)
			uniqueRequestsCount = 0
			mu.Unlock()

			logger.Printf("Unique requests in the last minute: %d", count)
		}
	}()

	if err := router.Run(":8080"); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
