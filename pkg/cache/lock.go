package cache

import (
	"context"
	"errors"
	"time"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/utils/idx"
)

type LockOption struct {
	Prefix string        // lock前缀
	Tries  int           // 加锁失败，尝试多少次
	TTL    time.Duration // 锁定时间，过期自动清除
	Wait   time.Duration // 每次加锁失败等待多久进行下次尝试
}

type Lock struct {
	opts  *LockOption
	cache *Cache
}

func NewLock(cache *Cache, opts *LockOption) *Lock {
	return &Lock{
		opts:  opts,
		cache: cache,
	}
}

// Acquire 获取锁
// 当前实现为了保证原子性加了一个全局的mutex，故在高并发场且锁粒度比较细的场景需要权衡性能问题
func (l *Lock) Acquire(ctx context.Context, target string) (identifier string, err error) {
	identifier = idx.UUIDv4()
	key := l.getKey(target)
	for i := 0; i < l.opts.Tries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			success, err := l.cache.SetNX(key, []byte(identifier), int(l.opts.TTL.Seconds()))
			if err != nil {
				return "", err
			}
			if success {
				return identifier, nil
			}
			time.Sleep(l.opts.Wait)
		}
	}
	return "", ecode.ErrLockFailed
}

func (l *Lock) Release(ctx context.Context, target, identifier string) (err error) {
	key := l.getKey(target)
	val, err := l.cache.Get(key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if string(val) == identifier {
		l.cache.Delete(key)
	}
	return nil
}

func (l *Lock) getKey(target string) string {
	return l.opts.Prefix + ":" + target
}
