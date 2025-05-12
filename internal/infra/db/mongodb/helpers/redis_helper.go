package helpers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClients     = make(map[string]*redis.Client)
	redisClientMutex sync.Mutex
)

var RedisTimeout = 30 * time.Second

func RedisHelper(connectionUrl string) *redis.Client {
	redisClientMutex.Lock()
	if client, exists := redisClients[connectionUrl]; exists {
		redisClientMutex.Unlock()
		return client
	}
	redisClientMutex.Unlock()

	opt, err := redis.ParseURL(connectionUrl)
	if err != nil {
		log.Fatalf("Error parsing Redis URL: %v", err)
		return nil
	}

	opt.PoolSize = 200
	opt.MinIdleConns = 20
	opt.ConnMaxIdleTime = 200 * time.Second

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Fatalf("Error pinging Redis: %v", err)
		return nil
	}

	redisClientMutex.Lock()
	redisClients[connectionUrl] = client
	redisClientMutex.Unlock()

	log.Printf("Connected to Redis: %s", connectionUrl)

	return client
}

func DisconnectRedis() {
	redisClientMutex.Lock()
	defer redisClientMutex.Unlock()

	for url, client := range redisClients {
		if err := client.Close(); err != nil {
			log.Printf("Error disconnecting from Redis %s: %v", url, err)
		} else {
			log.Printf("Disconnected from Redis: %s", url)
		}
	}

	redisClients = make(map[string]*redis.Client)
}
