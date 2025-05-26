package protocols

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// IMAPBrute IMAP爆破
func IMAPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	address := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 连接到IMAP服务器
	var conn net.Conn
	var err error

	if item.Port == 993 { // IMAPS
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

	// 发送LOGIN命令
	loginCmd := fmt.Sprintf("A001 LOGIN %s %s\r\n", item.Username, item.Password)
	_, err = conn.Write([]byte(loginCmd))
	if err != nil {
		result.Error = err
		return result
	}

	// 读取响应
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = err
		return result
	}

	response := string(buffer[:n])
	if strings.Contains(response, "A001 OK") {
		result.Success = true
		result.Banner = "IMAP login successful"
	} else {
		result.Error = fmt.Errorf("IMAP login failed: %s", strings.TrimSpace(response))
	}

	return result
}
