package protocols

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// POP3Brute POP3爆破
func POP3Brute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	address := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 连接到POP3服务器
	var conn net.Conn
	var err error

	if item.Port == 995 { // POP3S
		conn, err = tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = net.DialTimeout("tcp", address, timeout)
	}

	if err != nil {
		result.Error = err
		return result
	}
	defer conn.Close()

	// 设置超时
	conn.SetDeadline(time.Now().Add(timeout))

	// 读取服务器欢迎信息
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	// 发送USER命令
	userCmd := fmt.Sprintf("USER %s\r\n", item.Username)
	_, err = conn.Write([]byte(userCmd))
	if err != nil {
		result.Error = err
		return result
	}

	// 读取USER响应
	_, err = conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	// 发送PASS命令
	passCmd := fmt.Sprintf("PASS %s\r\n", item.Password)
	_, err = conn.Write([]byte(passCmd))
	if err != nil {
		result.Error = err
		return result
	}

	// 读取PASS响应
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	response := string(buffer[:n])
	if strings.HasPrefix(response, "+OK") {
		result.Success = true
		result.Banner = "POP3 login successful"
	} else {
		result.Error = fmt.Errorf("POP3 login failed: %s", strings.TrimSpace(response))
	}

	return result
}
