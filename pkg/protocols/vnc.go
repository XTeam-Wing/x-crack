package protocols

import (
	"context"
	"fmt"
	"net"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/mitchellh/go-vnc"
)

// VNCBrute VNC爆破
func VNCBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	address := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 使用上下文控制的连接
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to VNC server: %w", err)
		return result
	}
	defer conn.Close()

	// 创建VNC客户端配置
	cfg := &vnc.ClientConfig{
		Auth: []vnc.ClientAuth{
			&vnc.PasswordAuth{Password: item.Password},
		},
		Exclusive: false,
	}

	// 使用goroutine和select来控制VNC握手超时
	type vncResult struct {
		client *vnc.ClientConn
		err    error
	}

	vncChan := make(chan vncResult, 1)
	go func() {
		client, err := vnc.Client(conn, cfg)
		vncChan <- vncResult{client: client, err: err}
	}()

	select {
	case vncRes := <-vncChan:
		if vncRes.err != nil {
			result.Error = fmt.Errorf("VNC authentication failed: %w", vncRes.err)
			return result
		}
		defer vncRes.client.Close()

		result.Success = true
		result.Banner = "VNC authentication successful"
		return result

	case <-ctx.Done():
		result.Error = fmt.Errorf("VNC connection timeout: %w", ctx.Err())
		return result
	}
}
