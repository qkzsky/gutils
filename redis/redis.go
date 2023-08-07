package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/qkzsky/gutils/config"
	"github.com/qkzsky/gutils/logger"
	"runtime"
	"strings"
	"sync"
	"time"
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
	MaxOpen int
	MaxIdle int
}

var (
	redisMap = map[string]*Client{}
	mu       sync.RWMutex
)

func InitRedis() {
	for _, section := range config.Section("redis").ChildSections() {
		var err error
		name := strings.TrimPrefix(section.Name(), "redis.")
		mu.Lock()

		// 判断是否支持keycenter
		var sid = section.Key("sid").String()
		if sid != "" {
			var oldPassword = section.Key("auth").String()
			section.Key("auth").SetValue(string(keycenter.DecryptSimple(sid, oldPassword)))
		}

		redisMap[name], err = NewRedis(redisConfig{
			Host:    section.Key("host").String(),
			Port:    section.Key("port").String(),
			Auth:    section.Key("auth").String(),
			MaxOpen: section.Key("max_open").MustInt(defaultPoolSize),
			MaxIdle: section.Key("max_idle").MustInt(defaultIdleSize),
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
		Network:      "tcp",
		Addr:         c.Host + ":" + c.Port,
		Password:     c.Auth,
		DialTimeout:  DefaultConnectTimeout,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		PoolSize:     c.MaxOpen,
		MinIdleConns: c.MaxIdle,
		IdleTimeout:  180 * time.Second,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("[redis] " + err.Error())
	}

	return &Client{client}, nil
}
