package protocols

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// SMTPBrute SMTP爆破，支持 context 超时控制
func SMTPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	// 使用 channel 来传递结果，支持 context 取消
	type smtpResult struct {
		success bool
		err     error
		banner  string
	}

	resultChan := make(chan smtpResult, 1)

	// 在 goroutine 中执行 SMTP 操作
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- smtpResult{
					success: false,
					err:     fmt.Errorf("SMTP operation panic: %v", r),
				}
			}
		}()

		address := fmt.Sprintf("%s:%d", item.Target, item.Port)

		// 尝试连接SMTP服务器
		client, err := smtp.Dial(address)
		if err != nil {
			resultChan <- smtpResult{
				success: false,
				err:     fmt.Errorf("SMTP dial failed: %w", err),
			}
			return
		}
		defer client.Close()

		// 检查是否支持STARTTLS
		if ok, _ := client.Extension("STARTTLS"); ok {
			config := &tls.Config{
				ServerName:         item.Target,
				InsecureSkipVerify: true,
			}
			if err := client.StartTLS(config); err != nil {
				// STARTTLS失败，继续使用明文连接
			}
		}

		// 尝试认证
		auth := smtp.PlainAuth("", item.Username, item.Password, item.Target)
		err = client.Auth(auth)
		if err != nil {
			resultChan <- smtpResult{
				success: false,
				err:     fmt.Errorf("SMTP auth failed: %w", err),
			}
			return
		}

		resultChan <- smtpResult{
			success: true,
			banner:  "SMTP authentication successful",
		}
	}()

	// 等待结果或超时
	select {
	case smtpRes := <-resultChan:
		result.Success = smtpRes.success
		result.Error = smtpRes.err
		result.Banner = smtpRes.banner
		return result

	case <-ctx.Done():
		result.Error = fmt.Errorf("SMTP operation timeout after %v: %w", item.Timeout, ctx.Err())
		return result

	case <-time.After(item.Timeout + time.Second*2):
		// 额外的安全超时，防止 context 失效
		result.Error = fmt.Errorf("SMTP operation hard timeout after %v", item.Timeout+time.Second*2)
		return result
	}
}
