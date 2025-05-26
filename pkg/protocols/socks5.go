package protocols

import (
	"fmt"
	"net"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"golang.org/x/net/proxy"
)

// SOCKS5Brute SOCKS5爆破
func SOCKS5Brute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	// 构建SOCKS5服务器地址
	socks5Addr := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 方法1: 使用golang.org/x/net/proxy包进行SOCKS5认证
	if item.Username != "" && item.Password != "" {
		// 有用户名密码的认证
		auth := &proxy.Auth{
			User:     item.Username,
			Password: item.Password,
		}

		// 创建SOCKS5代理拨号器
		dialer, err := proxy.SOCKS5("tcp", socks5Addr, auth, &net.Dialer{
			Timeout: item.Timeout,
		})
		if err != nil {
			result.Error = fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
			return result
		}

		// 尝试通过代理连接到一个目标地址来验证认证
		testAddr := "223.5.5.5:53" // 使用Google DNS作为测试目标
		conn, err := dialer.Dial("tcp", testAddr)
		if err != nil {
			// 认证失败或连接失败
			result.Error = err
			return result
		}
		defer conn.Close()

		// 如果能成功建立连接，说明认证成功
		result.Success = true
		result.Banner = fmt.Sprintf("SOCKS5 authentication successful for %s:%s", item.Username, item.Password)

	} else {
		// 无认证的SOCKS5代理测试
		dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, &net.Dialer{
			Timeout: item.Timeout,
		})
		if err != nil {
			result.Error = fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
			return result
		}

		// 尝试通过代理连接
		testAddr := "8.8.8.8:53"
		conn, err := dialer.Dial("tcp", testAddr)
		if err != nil {
			result.Error = err
			return result
		}
		defer conn.Close()

		result.Success = true
		result.Banner = "SOCKS5 proxy connection successful (no auth)"
	}

	return result
}
