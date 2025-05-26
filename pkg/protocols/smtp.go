package protocols

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// SMTPBrute SMTP爆破
func SMTPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	address := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 尝试连接SMTP服务器
	client, err := smtp.Dial(address)
	if err != nil {
		result.Error = err
		return result
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
		result.Error = err
		return result
	}

	result.Success = true
	result.Banner = "SMTP authentication successful"
	return result
}
