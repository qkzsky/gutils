package redis

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/qkzsky/gutils/config"
	"github.com/qkzsky/gutils/logger"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultConnectTimeout = 100 * time.Millisecond
	DefaultReadTimeout    = 1000 * time.Millisecond
	DefaultWriteTimeout   = 1000 * time.Millisecond

	Nil = redis.Nil
)

var (
	defaultIdleSize = runtime.NumCPU() + 1
	defaultPoolSize = runtime.NumCPU()*2 + 1
)

type Client struct {
	*redis.Client
}

type redisConfig struct {
	Host    string
	Port    string
	Auth    string
	DB      int
	MaxOpen int
	MaxIdle int
}

var (
	redisMap = map[string]*Client{}
	mu       sync.RWMutex
)

func InitRedis() {
	redisConfigMap := config.GetStringMap("redis")

	for name, redisConf := range redisConfigMap {
		confMap, ok := redisConf.(map[string]interface{})
		if !ok {
			continue
		}

		mu.Lock()
		var err error
		redisMap[name], err = NewRedis(redisConfig{
			Host:    getStringFromMap(confMap, "host"),
			Port:    getStringFromMap(confMap, "port"),
			Auth:    getStringFromMap(confMap, "auth"),
			DB:      getIntFromMapWithDefault(confMap, "db", 0),
			MaxOpen: getIntFromMapWithDefault(confMap, "max_open", defaultPoolSize),
			MaxIdle: getIntFromMapWithDefault(confMap, "max_idle", defaultIdleSize),
		})
		mu.Unlock()

		if err != nil {
			panic(fmt.Sprintf("redis init failed. name: %s, error: %s.", name, err.Error()))
		}
	}
}

func GetRedis(name string) *Client {
	mu.RLock()
	defer mu.RUnlock()
	if client, ok := redisMap[name]; ok {
		return client
	}

	panic("redis not found: " + name)
}

func NewRedis(c redisConfig) (*Client, error) {
	if c.Host == "" || c.Port == "" {
		return nil, errors.New("host or port is empty")
	}

	client := redis.NewClient(&redis.Options{
		Network:         "tcp",
		Addr:            c.Host + ":" + c.Port,
		Password:        c.Auth,
		DB:              c.DB,
		DialTimeout:     DefaultConnectTimeout,
		ReadTimeout:     DefaultReadTimeout,
		WriteTimeout:    DefaultWriteTimeout,
		PoolSize:        c.MaxOpen,
		MinIdleConns:    c.MaxIdle,
		ConnMaxIdleTime: 180 * time.Second,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("[redis] " + err.Error())
	}

	return &Client{client}, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func getIntFromMapWithDefault(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}
