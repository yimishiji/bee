package db

import (
	"time"

	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

type redisClient struct {
	baseRedisClient *redis.Client
	prefix          string
}

// NewClient returns a client to the Redis Server specified by Options.
func NewClient(redisOption *redis.Options) *redisClient {
	client := redis.NewClient(redisOption)
	c := redisClient{
		baseRedisClient: client,
		prefix:          beego.AppConfig.String("redis::prefix"),
	}
	return &c
}

// Redis `GET key` command. It returns redis.Nil error when key does not exist.
func (c *redisClient) Get(key string) *redis.StringCmd {
	key = c.BuildKey(key)
	return c.baseRedisClient.Get(key)
}

func (c *redisClient) GetSet(key string, value interface{}) *redis.StringCmd {
	key = c.BuildKey(key)
	return c.baseRedisClient.GetSet(key, value)
}

// Redis `SET key value [expiration]` command.
//
// Use expiration for `SETEX`-like behavior.
// Zero expiration means the key has no expiration time.
func (c *redisClient) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	key = c.BuildKey(key)

	if _, isstring := value.(string); !isstring {
		if listStr, err := json.Marshal(value); err == nil {
			return c.baseRedisClient.Set(key, listStr, expiration)
		}
	}

	return c.baseRedisClient.Set(key, value, expiration)
}

//删除缓存
func (c *redisClient) Del(key ...string) {
	for i, _ := range key {
		key[i] = c.BuildKey(key[i])
	}
	c.baseRedisClient.Del(key...)
}

//主键加前缀
func (c *redisClient) BuildKey(key string) string {
	return c.prefix + key
}
