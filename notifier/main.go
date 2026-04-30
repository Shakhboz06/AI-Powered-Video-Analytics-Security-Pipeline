package main

import (
	"log"
	"video-analytics-pipe/config"
	"video-analytics-pipe/db/redis/cache"
)

func main() {

	redisAddr := config.GetString("REDIS_URL", "")
	redisPass := config.GetString("REDIS_PASS", "")
	redisDB := config.GetInt("REDIS_DB", 0)

	if redisAddr == "" {
		log.Fatal("REDIS_URL environment variable is required")
	}

	rdb := cache.NewRedisClient(redisAddr, redisPass, redisDB)
	defer rdb.Close()
	
	SubToRedis(rdb)

}