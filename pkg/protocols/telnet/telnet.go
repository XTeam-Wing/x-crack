package telnet

import (
	"context"
	"fmt"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// TelnetBrute Telnet爆破，支持 context 超时控制
func TelnetBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	// 使用 channel 来传递结果，支持 context 取消
	type telnetResult struct {
		success bool
		err     error
		banner  string
	}

	resultChan := make(chan telnetResult, 1)

	// 在 goroutine 中执行 Telnet 操作
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- telnetResult{
					success: false,
					err:     fmt.Errorf("Telnet operation panic: %v", r),
				}
			}
		}()

		client := New(item.Target, item.Port, item.Timeout)
		err := client.Connect()
		if err != nil {
			resultChan <- telnetResult{
				success: false,
				err:     fmt.Errorf("Telnet connect failed: %w", err),
			}
			return
		}
		defer client.Close()

		client.UserName = item.Username
		client.Password = item.Password
		client.ServerType = getTelnetServerType(item.Target, item.Port, item.Timeout)
		err = client.Login()
		if err != nil {
			resultChan <- telnetResult{
				success: false,
				err:     fmt.Errorf("Telnet login failed: %w", err),
			}
			return
		}

		resultChan <- telnetResult{
			success: true,
			banner:  "Telnet login successful",
		}
	}()

	// 等待结果或超时
	select {
	case telnetRes := <-resultChan:
		result.Success = telnetRes.success
		result.Error = telnetRes.err
		result.Banner = telnetRes.banner
		return result

	case <-ctx.Done():
		result.Error = fmt.Errorf("Telnet operation timeout after %v: %w", item.Timeout, ctx.Err())
		return result

	case <-time.After(item.Timeout + time.Second*2):
		// 额外的安全超时，防止 context 失效
		result.Error = fmt.Errorf("Telnet operation hard timeout after %v", item.Timeout+time.Second*2)
		return result
	}
}

func getTelnetServerType(ip string, port int, timeout time.Duration) int {
	client := New(ip, port, timeout)
	err := client.Connect()
	if err != nil {
		return Closed
	}
	defer client.Close()
	return client.MakeServerType()
}
