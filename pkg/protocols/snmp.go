package protocols

import (
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/gosnmp/gosnmp"
)

// SNMPBrute SNMP爆破
func SNMPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	// SNMP使用community字符串，通常作为"密码"传入
	community := item.Password
	if community == "" {
		community = "public" // 默认community
	}

	timeout := item.Timeout

	// 创建SNMP客户端
	g := &gosnmp.GoSNMP{
		Target:    item.Target,
		Port:      uint16(item.Port),
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   timeout,
		Retries:   1,
	}

	err := g.Connect()
	if err != nil {
		result.Error = err
		return result
	}
	defer g.Conn.Close()

	// 尝试获取系统信息 (sysDescr OID: 1.3.6.1.2.1.1.1.0)
	oids := []string{"1.3.6.1.2.1.1.1.0"}
	response, err := g.Get(oids)
	if err != nil {
		result.Error = err
		return result
	}

	if len(response.Variables) > 0 {
		result.Success = true
		result.Banner = fmt.Sprintf("SNMP community '%s' successful", community)
	} else {
		result.Error = fmt.Errorf("SNMP community '%s' failed", community)
	}

	return result
}
