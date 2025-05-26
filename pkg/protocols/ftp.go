package protocols

import (
	"fmt"
	"net"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// FTPBrute FTP爆破
func FTPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	target := fmt.Sprintf("%s:%d", item.Target, item.Port)
	conn, err := net.DialTimeout("tcp", target, item.Timeout)
	if err != nil {
		result.Error = err
		return result
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(item.Timeout))

	// 简单的FTP验证逻辑（这里需要实现完整的FTP协议）
	// 发送用户名
	_, err = conn.Write([]byte(fmt.Sprintf("USER %s\r\n", item.Username)))
	if err != nil {
		result.Error = err
		return result
	}

	// 发送密码
	_, err = conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", item.Password)))
	if err != nil {
		result.Error = err
		return result
	}

	// 读取响应
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	response := string(buffer[:n])
	if len(response) > 0 && response[0] == '2' {
		result.Success = true
		result.Banner = "FTP login successful"
	}

	return result
}
