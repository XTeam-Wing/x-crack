package protocols

import (
	"context"
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/go-redis/redis/v8"
)

// RedisBrute Redis爆破
func RedisBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username != "" {
		// Redis一般不使用用户名认证
		return result
	}

	timeout := item.Timeout

	// Redis连接配置
	rdb := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", item.Target, item.Port),
		Password:    item.Password,
		DB:          0,
		DialTimeout: timeout,
		ReadTimeout: timeout / 2,
		MaxRetries:  1,
	})
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 尝试连接并执行PING命令
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Banner = "Redis connection successful"
	return result
}
