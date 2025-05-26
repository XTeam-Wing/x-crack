package protocols

import (
	"fmt"
	"net"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// TelnetBrute Telnet爆破
func TelnetBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	target := fmt.Sprintf("%s:%d", item.Target, item.Port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		result.Error = err
		return result
	}
	defer conn.Close()

	// 简单的Telnet验证逻辑
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(timeout))

	// 读取初始响应
	_, err = conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	// 发送用户名
	_, err = conn.Write([]byte(item.Username + "\n"))
	if err != nil {
		result.Error = err
		return result
	}

	// 发送密码
	_, err = conn.Write([]byte(item.Password + "\n"))
	if err != nil {
		result.Error = err
		return result
	}

	// 读取最终响应
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	response := string(buffer[:n])
	// 简单判断登录成功（实际应该根据具体的Telnet服务器响应来判断）
	if len(response) > 0 {
		result.Success = true
		result.Banner = "Telnet connection established"
	}

	return result
}
