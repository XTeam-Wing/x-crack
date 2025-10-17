package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/gosnmp/gosnmp"
)

// SNMPBrute SNMP爆破，支持 context 超时控制
func SNMPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username != "" {
		return result
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	// SNMP使用community字符串，通常作为"密码"传入
	community := item.Password
	if community == "" {
		community = "public" // 默认community
	}

	// 使用 channel 来传递结果，支持 context 取消
	type snmpResult struct {
		success bool
		err     error
		banner  string
	}

	resultChan := make(chan snmpResult, 1)

	// 在 goroutine 中执行 SNMP 操作
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- snmpResult{
					success: false,
					err:     fmt.Errorf("SNMP operation panic: %v", r),
				}
			}
		}()

		// 创建SNMP客户端
		g := &gosnmp.GoSNMP{
			Target:    item.Target,
			Port:      uint16(item.Port),
			Community: community,
			Version:   gosnmp.Version2c,
			Timeout:   item.Timeout / 2, // 使用一半的超时时间
			Retries:   1,
		}

		err := g.Connect()
		if err != nil {
			resultChan <- snmpResult{
				success: false,
				err:     fmt.Errorf("SNMP connect failed: %w", err),
			}
			return
		}
		defer g.Conn.Close()

		// 尝试获取系统信息 (sysDescr OID: 1.3.6.1.2.1.1.1.0)
		oids := []string{"1.3.6.1.2.1.1.1.0"}
		response, err := g.Get(oids)
		if err != nil {
			resultChan <- snmpResult{
				success: false,
				err:     fmt.Errorf("SNMP get failed: %w", err),
			}
			return
		}

		if len(response.Variables) > 0 {
			resultChan <- snmpResult{
				success: true,
				banner:  fmt.Sprintf("SNMP community '%s' successful", community),
			}
		} else {
			resultChan <- snmpResult{
				success: false,
				err:     fmt.Errorf("SNMP community '%s' failed", community),
			}
		}
	}()

	// 等待结果或超时
	select {
	case snmpRes := <-resultChan:
		result.Success = snmpRes.success
		result.Error = snmpRes.err
		result.Banner = snmpRes.banner
		return result

	case <-ctx.Done():
		result.Error = fmt.Errorf("SNMP operation timeout after %v: %w", item.Timeout, ctx.Err())
		return result

	case <-time.After(item.Timeout + time.Second*2):
		// 额外的安全超时，防止 context 失效
		result.Error = fmt.Errorf("SNMP operation hard timeout after %v", item.Timeout+time.Second*2)
		return result
	}
}
