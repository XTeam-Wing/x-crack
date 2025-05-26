package protocols

import (
	"context"
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/yaklang/yaklang/common/utils/bruteutils"
)

// POP3Brute POP3爆破
func POP3Brute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	timeout := item.Timeout
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ok, err := bruteutils.POP3Auth(fmt.Sprintf("%s:%d", item.Target, item.Port), item.Username, item.Password, true)
	if err != nil {
		result.Error = err
		return result
	}
	result.Success = ok
	return result
}
