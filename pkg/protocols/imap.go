package protocols

import (
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/yaklang/yaklang/common/utils/bruteutils"
)

// IMAPBrute IMAP爆破
func IMAPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username == "" {
		return result
	}

	ok, err := bruteutils.IMAPAuth(fmt.Sprintf("%s:%d", item.Target, item.Port), item.Username, item.Password)
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = ok
	return result
}
