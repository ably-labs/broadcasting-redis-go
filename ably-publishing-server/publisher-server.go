package main

import (
	"log"
	"os"
	"time"
	"fmt"
	"context"
	"sync"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/ably/ably-go/ably"
)

var ctx = context.Background()

var wg = &sync.WaitGroup{}

func main() {
	client := getRedis()

	channel := getAblyChannel()

	go func() {
		for {
			transactionWithRedis(client, channel)
		}
	}()

	wg.Add(1)
	wg.Wait()
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

func transactionWithRedis(client *redis.Client, channel ably.RealtimeChannel) error {
	// Redis key where messages from the trading server are stored
	redisQueueName := getEnv("QUEUE_KEY", "myJobQueue")

	// Values to be used for checking our Redis log key for the rate limit
	redisLogName := redisQueueName + ":log"
	now := time.Now().UnixNano()
	windowSize := int64(time.Second)
	clearBefore := now - windowSize
	rateLimit, _ := strconv.ParseInt(getEnv("RATE_LIMIT", "50"), 10, 64)


	err := client.Watch(func(tx *redis.Tx) error {
		tx.ZRemRangeByScore(redisLogName, "0", strconv.FormatInt(clearBefore, 10))

		// Get the number of messages sent this second
		messagesThisSecond, err := tx.ZCard(redisLogName).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		// If under rate limit, indicate that we'll be publishing another message
		// And publish it to Ably
		if messagesThisSecond < rateLimit {
			err = tx.ZAdd(redisLogName, redis.Z{
				Score:  float64(now),
				Member: now,
			}).Err()
			if err != nil && err != redis.Nil {
				return err
			}

			messageToPublish, err := tx.BLPop(0*time.Second, redisQueueName).Result()
			if err != nil && err != redis.Nil {
				return err
			}

			_, err = channel.Publish("trade", messageToPublish[1])
			if err != nil {
				fmt.Println(err)
			}
		}

		return err
	}, redisLogName)

	return err
}

func getAblyChannel() ably.RealtimeChannel {
	opts := &ably.ClientOptions{
		AuthOptions: ably.AuthOptions{
			// If you have an Ably account, you can find
			// your API key at https://www.ably.io/accounts/any/apps/any/app_keys
			Key: getEnv("ABLY_KEY", "No key specified"),
		},
		// NoEcho:   true, // Uncomment to stop messages you send from being sent back
	}

	// Connect to Ably using the API key and ClientID specified above
	ablyClient, err := ably.NewRealtimeClient(opts)
	if err != nil {
		panic(err)
	}

	// Connect to the Ably Channel with name 'trades'
	return *ablyClient.Channels.Get(getEnv("CHANNEL_NAME", "trades"))
}

func getEnv(envName, valueDefault string) string {
	value := os.Getenv(envName)
	if value == "" {
		return valueDefault
	}
	return value
}
