package Lock

import (
	"context"
	_ "embed"
	"errors"
	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

type Client struct {
	client redis.Cmdable
	//redis.ClusterClient 集群
	//redis.Client 单机
	s singleflight.Group
}

var (
	ErrLockNotHold         = errors.New("未持有锁")
	ErrFailedTOPreemptLock = errors.New("加锁失败")
)

var (
	//go:embed unlock.lua
	luaUnlock string
	//go:embed refresh.lua
	luaRefresh string
	//go:embed lock.lua
	luaLock string
)

func NewClient(client redis.Cmdable) *Client {
	return &Client{client: client}
}
func (c *Client) SinglefightLock(ctx context.Context, key string,
	expiration time.Duration, retry RetryStrategy, timeout time.Duration) (*Lock, error) {

	for {
		flag := false
		resCh := c.s.DoChan(key, func() (interface{}, error) {
			flag = true
			return c.Lock(ctx, key, expiration, retry, timeout)
		})
		select {
		case res := <-resCh:
			if flag {
				if res.Err != nil {
					if res.Err != nil {
						return nil, res.Err
					}
					return res.Val.(*Lock), nil
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
func (c *Client) Lock(ctx context.Context, key string,
	expiration time.Duration, retry RetryStrategy, timeout time.Duration) (*Lock, error) {
	value := uuid.New().String()
	//计时器
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, value, expiration).Bool()
		cancel()

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		//如果是超时
		if res {
			return newLock(c.client, key, value, expiration), nil
		}
		interval, ok := retry.Next()
		if !ok {
			return nil, ErrFailedTOPreemptLock
		}
		//重试间隔，计时器
		if timer != nil {
			timer = time.NewTimer(interval)
		}
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:

		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	value := uuid.New().String()
	res, err := c.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !res {
		return nil, ErrFailedTOPreemptLock
	}
	return newLock(c.client, key, value, expiration), nil
}

type Lock struct {
	client     redis.Cmdable
	Key        string
	Value      string
	expiration time.Duration
	unlock     chan struct{}
}

func newLock(client redis.Cmdable, key string, value string, expiration time.Duration) *Lock {
	return &Lock{
		client:     client,
		Key:        key,
		Value:      value,
		expiration: expiration,
		unlock:     make(chan struct{}, 1),
	}
}
func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	ch := make(chan struct{}, 1)
	defer close(ch)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ch:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				//立即重试
				ch <- struct{}{}
				continue
			}
			if err != nil { //始终需要处理
				return err
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				//立即重试
				ch <- struct{}{}
				continue
			}
			if err != nil { //始终需要处理
				return err
			}
		case <-l.unlock:
			return nil

		}
	}
}
func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.Key}, l.Value, l.expiration.Milliseconds()).Int64()
	if redis.Nil == err {
		return ErrLockNotHold
	}
	if err != nil {
		return err
	}
	//判断res是否是1
	if res == 1 {
		//这个锁不存在或者不是你的
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) UnLock(ctx context.Context) error {
	//确保是这把锁
	/*



		val, err := l.client.Get(ctx,l.key).Result()
		if err != nil {
			return err
		}
		if val == l.value{
			_, err := l.client.Del(ctx, l.key).Result()
			if err != nil {
				return err
			}
		}

		//if res != 1 {
		//过期，被删
		//	return errors.New("解锁失败")
		//}
		return nil

	*/

	res, err := l.client.Eval(ctx, luaUnlock, []string{l.Key}, l.Value).Int64()
	defer func() {
		l.unlock <- struct{}{}
		close(l.unlock)
	}()
	if redis.Nil == err {
		return ErrLockNotHold
	}
	if err != nil {
		return err
	}
	//判断res是否是1
	if res == 0 {
		//这个锁不存在或者不是你的
		return ErrLockNotHold
	}
	return nil
}
