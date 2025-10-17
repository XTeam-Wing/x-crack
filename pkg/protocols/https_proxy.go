package protocols

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// HTTPSProxyBrute HTTP代理爆破
func HTTPSProxyBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	var httpProxyAddress string
	if item.Username == "" && item.Password == "" {
		httpProxyAddress = fmt.Sprintf("https://%s:%d", item.Target, item.Port)
	} else if item.Password != "" && item.Username != "" {
		httpProxyAddress = fmt.Sprintf("https://%s:%s@%s:%d", item.Username, item.Password, item.Target, item.Port)
	} else {
		return result
	}
	// 实现HTTP代理的验证逻辑
	proxyURL, err := url.Parse(httpProxyAddress)
	if err != nil {
		result.Error = err
		return result
	}
	httpTransport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: httpTransport,
	}

	// 例如使用http.Client发送请求，设置代理地址等
	req, err := http.NewRequest("GET", "https://baidu.com", nil)
	if err != nil {
		result.Error = err
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode == http.StatusOK {
		result.Success = true
		result.Banner = "HTTP Proxy authentication successful"
	}

	return result
}
