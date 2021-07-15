package main

import (
	"log"
	"os"
	"fmt"
	"sync"
	"time"
	"math/rand"
	"strconv"

	"github.com/go-redis/redis"
)

var wg = &sync.WaitGroup{}

func main() {
	client := getRedis()

	go publishingLoop(client)

	// Add an item to the wait group so the server keeps running
	wg.Add(1)
	wg.Wait()
}

func getEnv(envName, valueDefault string) string {
	value := os.Getenv(envName)
	if value == "" {
		return valueDefault
	}
	return value
}

func getRedis() *redis.Client {
	// Create Redis Client
	var (
		host     = getEnv("REDIS_HOST", "localhost")
		port     = string(getEnv("REDIS_PORT", "6379"))
		password = getEnv("REDIS_PASSWORD", "")
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func publishingLoop(redisClient *redis.Client) {
	queueKey := getEnv("QUEUE_KEY", "myJobQueue")
	publishRate, _ := strconv.Atoi(getEnv("PUBLISH_RATE", "200"))
	baseTradeValue := float64(200)

	// Send a burst of messages to Redis every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <- ticker.C:
			for i := 0; i < publishRate; i++ {
				// Random test value varying +- 5 around the baseTradeValue
				tradeValue := baseTradeValue + (rand.Float64() * 10 - 5)
				redisClient.RPush(queueKey, fmt.Sprintf("%f", tradeValue))
			}
		case <- quit:
			// If you want a way to cleanly stop the server, call quit <- true
			// So this runs
			ticker.Stop()
			defer wg.Done()
			return
		}
	}
}
