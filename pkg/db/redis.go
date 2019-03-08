package db

import (
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

var (
	Redis *RedisClient
)

//连接redis
func GetRedisClient() (*RedisClient, error) {
	db, _ := beego.AppConfig.Int("redis::database")
	redisOption := &redis.Options{
		Addr:     beego.AppConfig.String("redis::host"),
		Password: beego.AppConfig.String("redis::password"),
		DB:       db,
	}
	client := NewClient(redisOption)

	_, err := client.baseRedisClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	Redis = client

	return client, nil
}
