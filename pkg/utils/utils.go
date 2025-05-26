package utils

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParsePortRange 解析端口范围
func ParsePortRange(portRange string) ([]int, error) {
	if portRange == "" {
		return []int{}, nil
	}

	var ports []int
	parts := strings.Split(portRange, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// 范围格式：1-1000
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start port: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end port: %s", rangeParts[1])
			}

			if start > end {
				start, end = end, start
			}

			for i := start; i <= end; i++ {
				if i >= 1 && i <= 65535 {
					ports = append(ports, i)
				}
			}
		} else {
			// 单个端口
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}
			if port >= 1 && port <= 65535 {
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

// LoadLinesFromFile 从文件加载行
func LoadLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

// IsPortOpen 检查端口是否开放
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// RandomDelay 随机延迟
func RandomDelay(min, max time.Duration) {
	if min >= max {
		time.Sleep(min)
		return
	}

	// 计算随机延迟
	diff := max - min
	randomDuration := time.Duration(rand.Int63n(int64(diff)))
	time.Sleep(min + randomDuration)
}

// ValidateTarget 验证目标格式
func ValidateTarget(target string) error {
	if target == "" {
		return fmt.Errorf("target cannot be empty")
	}

	// 检查是否为有效的IP地址或域名
	if net.ParseIP(target) != nil {
		return nil
	}

	// 简单的域名验证
	if strings.Contains(target, ".") && !strings.Contains(target, " ") {
		return nil
	}

	return fmt.Errorf("invalid target format: %s", target)
}

// ValidatePort 验证端口范围
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

// FilterPorts 过滤端口列表
func FilterPorts(ports []int, excludePorts []int) []int {
	excludeMap := make(map[int]bool)
	for _, port := range excludePorts {
		excludeMap[port] = true
	}

	var filtered []int
	for _, port := range ports {
		if !excludeMap[port] {
			filtered = append(filtered, port)
		}
	}

	return filtered
}

// GetDefaultPorts 获取协议的默认端口
func GetDefaultPorts(protocol string) []int {
	defaults := map[string][]int{
		"ssh":        {22},
		"ftp":        {21},
		"telnet":     {23},
		"mysql":      {3306},
		"postgresql": {5432},
		"redis":      {6379},
		"mongodb":    {27017},
		"http":       {80, 8080, 8000, 8888},
		"https":      {443, 8443},
		"smb":        {445, 139},
		"rdp":        {3389},
		"vnc":        {5900, 5901, 5902},
		"snmp":       {161},
		"imap":       {143, 993},
		"pop3":       {110, 995},
		"smtp":       {25, 587, 465},
	}

	if ports, exists := defaults[strings.ToLower(protocol)]; exists {
		return ports
	}

	return []int{}
}

// GenerateCommonUsernames 生成常见用户名
func GenerateCommonUsernames() []string {
	return []string{
		"admin", "administrator", "root", "user", "test", "guest", "oracle", "postgres",
		"mysql", "mssql", "sa", "ftp", "mail", "email", "web", "www", "http", "tomcat",
		"jenkins", "git", "svn", "redis", "mongodb", "elastic", "kibana", "grafana",
		"nagios", "zabbix", "cacti", "pi", "ubuntu", "centos", "debian", "redhat",
		"service", "daemon", "nobody", "www-data", "apache", "nginx", "operator",
	}
}

// GenerateCommonPasswords 生成常见密码
func GenerateCommonPasswords() []string {
	return []string{
		"123456", "password", "123456789", "12345678", "12345", "1234567890",
		"1234567", "password123", "000000", "123123", "admin", "admin123",
		"root", "pass", "test", "guest", "123", "1234", "12345", "123456",
		"password1", "123qwe", "qwerty", "abc123", "Password1", "welcome",
		"login", "changeme", "secret", "administrator", "letmein", "dragon",
		"master", "hello", "freedom", "whatever", "qazwsx", "trustno1",
		"", "admin", "root", "guest", "test", "oracle", "postgres", "mysql",
		"sa", "operator", "manager", "service", "support", "user", "demo",
	}
}

// ShuffleStrings 打乱字符串切片
func ShuffleStrings(slice []string) []string {
	shuffled := make([]string, len(slice))
	copy(shuffled, slice)

	rand.Seed(time.Now().UnixNano())
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// DeduplicateStrings 去重字符串切片
func DeduplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}
