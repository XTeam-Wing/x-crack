package protocols

import (
	"context"
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/XTeam-Wing/x-crack/pkg/protocols/grdp"
)

// RDPBrute RDP爆破
func RDPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username == "" {
		return result
	}

	target := fmt.Sprintf("%s:%d", item.Target, item.Port)

	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	var err error

	// 检查协议类型并尝试登录，使用 goroutine + select 模式处理超时
	protocolChan := make(chan string, 1)
	go func() {
		protocolChan <- grdp.VerifyProtocol(target)
	}()

	var protocol string
	select {
	case protocol = <-protocolChan:
		// 正常获取到协议类型
	case <-ctx.Done():
		// 超时或取消
		result.Error = fmt.Errorf("RDP protocol verification timeout: %w", ctx.Err())
		return result
	}

	fmt.Println("Detected protocol:", protocol)
	if protocol == grdp.PROTOCOL_SSL {
		// 需要检查grdp库是否支持上下文，如果不支持，使用goroutine+select模式
		errChan := make(chan error, 1)
		go func() {
			errChan <- grdp.LoginForSSL(target, item.Target, item.Username, item.Password)
		}()

		select {
		case err = <-errChan:
			// 正常完成
		case <-ctx.Done():
			// 超时或取消
			err = ctx.Err()
		}
	} else {
		// 同样的模式处理RDP
		errChan := make(chan error, 1)
		go func() {
			errChan <- grdp.LoginForRDP(target, item.Target, item.Username, item.Password)
		}()

		select {
		case err = <-errChan:
			// 正常完成
		case <-ctx.Done():
			// 超时或取消
			err = ctx.Err()
		}
	}

	if err != nil {
		result.Error = fmt.Errorf("RDP connection failed: %w", err)
		return result
	}

	result.Success = true
	result.Banner = "RDP connection successful"
	return result
}
