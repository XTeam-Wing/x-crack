package protocols

import (
	"context"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/yaklang/yaklang/common/utils/bruteutils"
)

// RDPBrute RDP爆破
func RDPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	_, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	ok, err := bruteutils.RDPLogin(item.Target, item.Target,
		item.Username, item.Password, item.Port)
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = ok
	return result
}
