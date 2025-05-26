package protocols

import (
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"golang.org/x/crypto/ssh"
)

// SSHBrute SSH爆破
func SSHBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	config := &ssh.ClientConfig{
		User: item.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(item.Password),
		},
		Timeout:         timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	target := fmt.Sprintf("%s:%d", item.Target, item.Port)
	client, err := ssh.Dial("tcp", target, config)
	if err != nil {
		result.Error = err
		return result
	}
	defer client.Close()

	result.Success = true
	result.Banner = "SSH connection successful"
	return result
}
