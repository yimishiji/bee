package db

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

type RedisClient struct {
	baseRedisClient *redis.Client
	prefix          string
}

// NewClient returns a client to the Redis Server specified by Options.
func NewClient(redisOption *redis.Options) *RedisClient {
	client := redis.NewClient(redisOption)
	c := RedisClient{
		baseRedisClient: client,
		prefix:          beego.AppConfig.String("redis::prefix"),
	}
	return &c
}

// Redis `GET key` command. It returns redis.Nil error when key does not exist.
func (c *RedisClient) Get(key string) *redis.StringCmd {
	key = c.BuildKey(key)
	return c.baseRedisClient.Get(key)
}

func (c *RedisClient) GetSet(key string, value interface{}) *redis.StringCmd {
	key = c.BuildKey(key)
	return c.baseRedisClient.GetSet(key, value)
}

// Redis `SET key value [expiration]` command.
//
// Use expiration for `SETEX`-like behavior.
// Zero expiration means the key has no expiration time.
func (c *RedisClient) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	key = c.BuildKey(key)
	return c.baseRedisClient.Set(key, value, expiration)
}

//主键加前缀
func (c *RedisClient) BuildKey(key string) string {
	return c.prefix + key
}
