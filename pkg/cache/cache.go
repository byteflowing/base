package cache

import (
	"errors"
	"sync"

	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	"github.com/coocood/freecache"
)

var (
	ErrNotFound = errors.New("not found")
)

type Cache struct {
	mux sync.Mutex

	cli *freecache.Cache
}

func New(opts *configv1.LocalCacheConfig) *Cache {
	return &Cache{
		cli: freecache.NewCache(int(opts.LocalCacheCapacity)),
	}
}

// Set sets a key, value and expiration for a cache entry and stores it in the cache.
// If the key is larger than 65535 or value is larger than 1/ 1024 of the cache size,
// the entry will not be written to the cache.
// expireSeconds <= 0 means no expire, but it can be evicted when cache is full
func (c *Cache) Set(key string, value []byte, expireSeconds int) (err error) {
	return c.cli.Set([]byte(key), value, expireSeconds)
}

// SetNX 当key不存在时添加
// 这里使用了全局的额mutex来保证原子性，故在key粒度比较细的高并发场景需要权衡性能
func (c *Cache) SetNX(key string, value []byte, expireSeconds int) (bool, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, err := c.cli.Get([]byte(key))
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, freecache.ErrNotFound) {
		return false, err
	}
	if err := c.cli.Set([]byte(key), value, expireSeconds); err != nil {
		return false, err
	}
	return true, nil
}

// Get 获取key对应的值
// 如果没有找到返回ErrNotFound
func (c *Cache) Get(key string) (value []byte, err error) {
	value, err = c.cli.Get([]byte(key))
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			err = ErrNotFound
		}
	}
	return
}

func (c *Cache) Exists(key string) (exist bool, err error) {
	_, err = c.cli.Get([]byte(key))
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Delete 删除key指定的值
func (c *Cache) Delete(key string) {
	c.cli.Del([]byte(key))
}
