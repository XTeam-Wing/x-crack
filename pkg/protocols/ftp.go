package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/jlaffaye/ftp"
)

// FTPBrute FTP爆破，支持 context 超时控制
func FTPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username == "" || item.Password == "" {
		return result
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	// 使用 channel 来传递结果，支持 context 取消
	type ftpResult struct {
		success bool
		err     error
		banner  string
	}

	resultChan := make(chan ftpResult, 1)

	// 在 goroutine 中执行 FTP 操作
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- ftpResult{
					success: false,
					err:     fmt.Errorf("FTP operation panic: %v", r),
				}
			}
		}()

		target := fmt.Sprintf("%s:%d", item.Target, item.Port)

		// 连接 FTP 服务器

		c, err := ftp.Dial(target, ftp.DialWithTimeout(item.Timeout))
		if err != nil {
			resultChan <- ftpResult{
				success: false,
				err:     fmt.Errorf("FTP dial failed: %w", err),
			}
			return
		}
		defer c.Quit()

		// 尝试登录
		err = c.Login(item.Username, item.Password)
		if err != nil {
			resultChan <- ftpResult{
				success: false,
				err:     fmt.Errorf("FTP login failed: %w", err),
			}
			return
		}

		// 登录成功，验证连接状态
		_, err = c.CurrentDir()
		if err != nil {
			resultChan <- ftpResult{
				success: false,
				err:     fmt.Errorf("FTP connection verification failed: %w", err),
			}
			return
		}

		resultChan <- ftpResult{
			success: true,
			banner:  "FTP login successful",
		}
	}()

	// 等待结果或超时
	select {
	case ftpRes := <-resultChan:
		result.Success = ftpRes.success
		result.Error = ftpRes.err
		result.Banner = ftpRes.banner
		return result

	case <-ctx.Done():
		result.Error = fmt.Errorf("FTP operation timeout after %v: %w", item.Timeout, ctx.Err())
		return result

	case <-time.After(item.Timeout + time.Second*2):
		// 额外的安全超时，防止 context 失效
		result.Error = fmt.Errorf("FTP operation hard timeout after %v", item.Timeout+time.Second*2)
		return result
	}
}
