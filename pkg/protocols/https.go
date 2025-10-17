package protocols

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// HTTPSBrute HTTPS基础认证爆破
func HTTPSBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username == "" || item.Password == "" {
		return result
	}
	timeout := item.Timeout
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := fmt.Sprintf("https://%s:%d/", item.Target, item.Port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = err
		return result
	}

	req.SetBasicAuth(item.Username, item.Password)

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	// HTTP 200/302/3xx 认为成功，401/403 认为失败
	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		result.Success = true
		result.Banner = fmt.Sprintf("HTTPS Basic Auth successful (Status: %d)", resp.StatusCode)
	} else {
		result.Error = fmt.Errorf("HTTPS Basic Auth failed (Status: %d)", resp.StatusCode)
	}

	return result
}
