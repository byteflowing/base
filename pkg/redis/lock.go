package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/utils/idx"
)

const unLockScript = `
	if
		redis.call("GET", KEYS[1]) == ARGV[1]
	then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`

const (
	keyFormat = "%s:{%s}"
)

type Lock struct {
	opts *LockOption

	rdb *Redis
	sha string
}

type LockOption struct {
	Prefix string        // lock前缀
	Tries  int           // 加锁失败，尝试多少次
	TTL    time.Duration // 锁定时间，过期自动清除
	Wait   time.Duration // 每次加锁失败等待多久进行下次尝试
}

func NewLock(rdb *Redis, opts *LockOption) *Lock {
	sha, err := rdb.ScriptLoad(context.Background(), unLockScript).Result()
	if err != nil {
		panic(err)
	}
	return &Lock{
		opts: opts,
		rdb:  rdb,
		sha:  sha,
	}
}

func (l *Lock) Acquire(ctx context.Context, target string) (identifier string, err error) {
	identifier = idx.UUIDv4()
	key := l.getKey(target)
	for i := 0; i < l.opts.Tries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			success, err := l.rdb.SetNX(ctx, key, identifier, l.opts.TTL).Result()
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
	keys := []string{key}
	res, err := l.rdb.EvalShaWithReload(ctx, l.sha, unLockScript, keys, identifier)
	if err != nil {
		return
	}
	result := res.(int64)
	if result == 1 {
		return nil
	}
	return ecode.ErrUnLockFailed
}

func (l *Lock) getKey(target string) string {
	return fmt.Sprintf(keyFormat, l.opts.Prefix, target)
}
