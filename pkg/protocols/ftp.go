package protocols

import (
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/jlaffaye/ftp"
)

// FTPBrute FTP爆破
func FTPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	target := fmt.Sprintf("%s:%d", item.Target, item.Port)
	c, err := ftp.Dial(target, ftp.DialWithTimeout(item.Timeout))
	if err != nil {
		result.Error = err
		return result
	}
	defer c.Quit()

	err = c.Login(item.Username, item.Password)
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Banner = "FTP login successful"
	return result
}
