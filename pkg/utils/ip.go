package utils

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// HasLocalIPAddr 检测 IPList 地址字符串是否是内网地址
func HasLocalIPAddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocalIP 检测 IPList 地址是否是内网地址
// 通过直接对比ip段范围效率更高
func HasLocalIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}

// IPToUInt32 将点分格式的IP地址转换为UINT32
func IPToUInt32(ip string) uint32 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum uint32
	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)

	return sum
}

func UInt32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func CheckIPV4(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])$`
	r, _ := regexp.Compile(ipReg)

	return r.MatchString(ip)
}

func CheckIPV4Subnet(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])/\d{1,2}$`
	r, _ := regexp.Compile(ipReg)

	return r.MatchString(ip)
}

func CheckIpRange(ip string) bool {
	address := strings.Split(ip, "-")
	if len(address) == 2 && CheckIPV4(address[0]) {
		return true
	}
	return false
}

func ParseIP(ip string) (ipAddressList []string) {
	// 192.168.1.1
	if CheckIPV4(ip) {
		return []string{ip}
	}
	// 192.168.1.0/24
	if CheckIPV4Subnet(ip) {
		addr, ipv4sub, err := net.ParseCIDR(ip)
		if err != nil {
			return
		}
		ones, bits := ipv4sub.Mask.Size()
		ipStart := IPToUInt32(addr.String())
		ipSize := int(math.Pow(2, float64(bits-ones)))
		for i := 0; i < ipSize; i++ {
			if i == 0 || i == 255 {
				continue
			}
			ipAddressList = append(ipAddressList, UInt32ToIP(uint32(i)+ipStart))
		}
		return
	}
	// 192.168.1.1-192.168.1.5
	if strings.Contains(ip, "-") && !strings.Contains(ip, " ") {
		address := strings.Split(ip, "-")
		if len(address) == 2 && CheckIPV4(address[0]) && CheckIPV4(address[1]) {
			ipStart := address[0]
			ipEnd := address[1]
			for i := IPToUInt32(ipStart); i <= IPToUInt32(ipEnd); i++ {
				if i == 0 || i == 255 {
					continue
				}
				ipAddressList = append(ipAddressList, UInt32ToIP(i))
			}
			return
		}
		if len(address) == 2 && CheckIPV4(address[0]) {
			ipStart := address[0]
			ipEnd := address[1]
			ipEnd = strings.Replace(ipEnd, " ", "", -1)
			ipEndInt, err := strconv.Atoi(ipEnd)
			if err != nil {
				return
			}
			ipStartInt := IPToUInt32(ipStart)
			for i := 0; i <= ipEndInt; i++ {
				if i == 0 || i == 255 {
					continue
				}
				ipAddressList = append(ipAddressList, UInt32ToIP(uint32(i)+ipStartInt))
			}
			return
		}
	}
	return
}

func CheckIP(ip string) bool {
	// 192.168.1.1
	if CheckIPV4(ip) || CheckIPV4Subnet(ip) || CheckIpRange(ip) {
		return true
	}

	return false
}
