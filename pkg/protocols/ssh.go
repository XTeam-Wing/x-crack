package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/projectdiscovery/gologger"
	"golang.org/x/crypto/ssh"
)

// SSHBrute SSH爆破，支持 context 超时控制
func SSHBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username == "" {
		return result
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	// 使用 channel 来传递结果，支持 context 取消
	type sshResult struct {
		success bool
		err     error
		banner  string
	}

	resultChan := make(chan sshResult, 1)

	// 在 goroutine 中执行 SSH 操作
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- sshResult{
					success: false,
					err:     fmt.Errorf("SSH operation panic: %v", r),
				}
			}
		}()

		config := &ssh.ClientConfig{
			User: item.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(item.Password),
			},
			Timeout:         item.Timeout / 2, // 给SSH内部操作一半的时间
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		target := fmt.Sprintf("%s:%d", item.Target, item.Port)
		gologger.Debug().Msgf("Attempting SSH connection to %s with user %s password %s", target, item.Username, item.Password)

		client, err := ssh.Dial("tcp", target, config)
		if err != nil {
			resultChan <- sshResult{
				success: false,
				err:     fmt.Errorf("SSH dial failed: %w", err),
			}
			return
		}
		defer client.Close()

		// 创建一个简单的session来验证连接
		session, err := client.NewSession()
		if err != nil {
			resultChan <- sshResult{
				success: false,
				err:     fmt.Errorf("SSH session creation failed: %w", err),
			}
			return
		}
		defer session.Close()

		resultChan <- sshResult{
			success: true,
			banner:  "SSH connection successful",
		}
	}()

	// 等待结果或超时
	select {
	case sshRes := <-resultChan:
		result.Success = sshRes.success
		result.Error = sshRes.err
		result.Banner = sshRes.banner
		return result

	case <-ctx.Done():
		result.Error = fmt.Errorf("SSH operation timeout after %v: %w", item.Timeout, ctx.Err())
		return result

	case <-time.After(item.Timeout + time.Second*2):
		// 额外的安全超时，防止 context 失效
		result.Error = fmt.Errorf("SSH operation hard timeout after %v", item.Timeout+time.Second*2)
		return result
	}
}
