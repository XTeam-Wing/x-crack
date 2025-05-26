package telnet

import (
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

// TelnetBrute Telnet爆破
func TelnetBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	client := New(item.Target, item.Port, item.Timeout)
	err := client.Connect()
	if err != nil {
		result.Error = err
		return result
	}
	defer client.Close()
	client.UserName = item.Username
	client.Password = item.Password
	client.ServerType = getTelnetServerType(item.Target, item.Port, item.Timeout)
	err = client.Login()
	if err != nil {
		result.Error = err
		return result
	}
	result.Success = true
	return result
}

func getTelnetServerType(ip string, port int, timeout time.Duration) int {
	client := New(ip, port, timeout)
	err := client.Connect()
	if err != nil {
		return Closed
	}
	defer client.Close()
	return client.MakeServerType()
}
