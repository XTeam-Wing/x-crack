package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ServiceTarget 服务目标结构
type ServiceTarget struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// ParseServiceURL 解析服务URL
// 支持格式：protocol://host:port, protocol://host (使用默认端口)
func ParseServiceURL(serviceURL string) (*ServiceTarget, error) {
	if serviceURL == "" {
		return nil, fmt.Errorf("empty service URL")
	}

	// 确保URL有协议前缀
	if !strings.Contains(serviceURL, "://") {
		return nil, fmt.Errorf("invalid service URL format: %s, expected format: protocol://host:port", serviceURL)
	}

	parsedURL, err := url.Parse(serviceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service URL '%s': %w", serviceURL, err)
	}

	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("missing protocol in service URL: %s", serviceURL)
	}

	if parsedURL.Host == "" {
		return nil, fmt.Errorf("missing host in service URL: %s", serviceURL)
	}

	target := &ServiceTarget{
		Protocol: strings.ToLower(parsedURL.Scheme),
	}

	// 处理主机和端口
	if parsedURL.Port() != "" {
		// 显式指定了端口
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port in service URL '%s': %w", serviceURL, err)
		}
		target.Host = parsedURL.Hostname()
		target.Port = port
	} else {
		// 没有指定端口，使用默认端口
		target.Host = parsedURL.Host
		defaultPorts := GetDefaultPorts(target.Protocol)
		if len(defaultPorts) > 0 {
			target.Port = defaultPorts[0] // 使用第一个默认端口
		} else {
			return nil, fmt.Errorf("no default port found for protocol '%s' in service URL: %s", target.Protocol, serviceURL)
		}
	}

	// 验证端口范围
	if err := ValidatePort(target.Port); err != nil {
		return nil, fmt.Errorf("invalid port in service URL '%s': %w", serviceURL, err)
	}

	return target, nil
}

// ParseServiceTargetFile 解析服务目标文件
// 文件格式每行一个服务URL，如：telnet://1.1.1.1:23
func ParseServiceTargetFile(filename string) ([]ServiceTarget, error) {
	lines, err := LoadLinesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load service target file '%s': %w", filename, err)
	}

	var targets []ServiceTarget
	for i, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue // 跳过空行和注释行
		}

		target, err := ParseServiceURL(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d in file '%s': %w", i+1, filename, err)
		}

		targets = append(targets, *target)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid service targets found in file: %s", filename)
	}

	return targets, nil
}

// ValidateServiceTarget 验证服务目标
func ValidateServiceTarget(target ServiceTarget) error {
	if target.Protocol == "" {
		return fmt.Errorf("empty protocol in service target")
	}

	if target.Host == "" {
		return fmt.Errorf("empty host in service target")
	}

	if err := ValidateTarget(target.Host); err != nil {
		return fmt.Errorf("invalid host in service target: %w", err)
	}

	if err := ValidatePort(target.Port); err != nil {
		return fmt.Errorf("invalid port in service target: %w", err)
	}

	return nil
}

// GroupServiceTargetsByProtocol 按协议分组服务目标
func GroupServiceTargetsByProtocol(targets []ServiceTarget) map[string][]ServiceTarget {
	groups := make(map[string][]ServiceTarget)

	for _, target := range targets {
		protocol := target.Protocol
		groups[protocol] = append(groups[protocol], target)
	}

	return groups
}

// GetUniqueProtocols 获取服务目标中的唯一协议列表
func GetUniqueProtocols(targets []ServiceTarget) []string {
	protocolSet := make(map[string]bool)
	for _, target := range targets {
		protocolSet[target.Protocol] = true
	}

	var protocols []string
	for protocol := range protocolSet {
		protocols = append(protocols, protocol)
	}

	return protocols
}

// GetUniqueHosts 获取服务目标中的唯一主机列表
func GetUniqueHosts(targets []ServiceTarget) []string {
	hostSet := make(map[string]bool)
	for _, target := range targets {
		hostSet[target.Host] = true
	}

	var hosts []string
	for host := range hostSet {
		hosts = append(hosts, host)
	}

	return hosts
}

// ConvertServiceTargetsToBruteTargets 将服务目标转换为爆破目标
func ConvertServiceTargetsToBruteTargets(targets []ServiceTarget) []struct {
	Type string
	Host string
	Port int
} {
	var bruteTargets []struct {
		Type string
		Host string
		Port int
	}

	for _, target := range targets {
		bruteTargets = append(bruteTargets, struct {
			Type string
			Host string
			Port int
		}{
			Type: target.Protocol,
			Host: target.Host,
			Port: target.Port,
		})
	}

	return bruteTargets
}
