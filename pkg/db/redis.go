package db

import (
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

var (
	Redis *redis.Client
)

//连接redis
func GetRedisClient() (*redis.Client, error) {
	db, _ := beego.AppConfig.Int("redis::database")
	redisOption := &redis.Options{
		Addr:     beego.AppConfig.String("redis::host"),
		Password: beego.AppConfig.String("redis::password"),
		DB:       db,
	}
	client := redis.NewClient(redisOption)

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	Redis = client

	return client, nil
}
