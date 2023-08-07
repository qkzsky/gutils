package cache

import (
	"github.com/coocood/freecache"
	"github.com/qkzsky/gutils/logger"
	"go.uber.org/zap"
)

var cache *freecache.Cache

//func init() {
//	section := config.Section("cache")
//	size, err := section.Key("size").Int()
//	if err != nil {
//		logger.Warn(err.Error())
//		return
//	}
//
//	cache = NewCache(size)
//	if size > 4 {
//		debug.SetGCPercent(20)
//	}
//}

func NewCache(size int) *freecache.Cache {
	// size 单位 MB
	if size < 4 {
		logger.Warn("memory cache size is too small", zap.Int("size", size))
		return nil
	}
	return freecache.NewCache(size << 20)
}

func GetCache() *freecache.Cache {
	return cache
}
