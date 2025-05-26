package protocols

import (
	"fmt"
	"net"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/hirochachacha/go-smb2"
)

// SMBBrute SMB爆破
func SMBBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	address := fmt.Sprintf("%s:%d", item.Target, item.Port)

	// 连接到SMB服务器
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.Error = err
		return result
	}
	defer conn.Close()

	// 创建SMB2客户端
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     item.Username,
			Password: item.Password,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		result.Error = err
		return result
	}
	defer s.Logoff()

	// 尝试连接到IPC$共享来验证认证
	fs, err := s.Mount("IPC$")
	if err != nil {
		result.Error = err
		return result
	}
	defer fs.Umount()

	result.Success = true
	result.Banner = "SMB authentication successful"
	return result
}
