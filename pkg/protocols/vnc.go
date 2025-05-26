package protocols

import (
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

	timeout := item.Timeout
	address := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 连接到VNC服务器
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.Error = err
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

	// 尝试连接VNC
	client, err := vnc.Client(conn, cfg)
	if err != nil {
		result.Error = err
		return result
	}
	defer client.Close()

	result.Success = true
	result.Banner = "VNC authentication successful"
	return result
}
