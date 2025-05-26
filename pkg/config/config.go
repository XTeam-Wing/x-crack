package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// AppConfig 应用配置
type AppConfig struct {
	// 基本设置
	Version  string `yaml:"version"`
	Debug    bool   `yaml:"debug"`
	LogLevel string `yaml:"log_level"`

	// 爆破设置
	Brute BruteConfig `yaml:"brute"`

	// 输出设置
	Output OutputConfig `yaml:"output"`

	// 代理设置
	Proxy ProxyConfig `yaml:"proxy"`
}

// BruteConfig 爆破配置
type BruteConfig struct {
	// 并发控制
	TargetConcurrent int `yaml:"target_concurrent"`
	TaskConcurrent   int `yaml:"task_concurrent"`

	// 延迟控制
	MinDelay string `yaml:"min_delay"`
	MaxDelay string `yaml:"max_delay"`

	// 超时设置
	Timeout string `yaml:"timeout"`

	// 重试设置
	MaxRetries int `yaml:"max_retries"`

	// 停止条件
	OkToStop           bool `yaml:"ok_to_stop"`
	FinishingThreshold int  `yaml:"finishing_threshold"`

	// 字典设置
	DefaultUserDict []string `yaml:"default_user_dict"`
	DefaultPassDict []string `yaml:"default_pass_dict"`

	// 其他设置
	SkipEmptyPassword  bool `yaml:"skip_empty_password"`
	SkipEmptyUsername  bool `yaml:"skip_empty_username"`
	OnlyNeedPassword   bool `yaml:"only_need_password"`
	AllowBlankUsername bool `yaml:"allow_blank_username"` // 允许空用户名
	AllowBlankPassword bool `yaml:"allow_blank_password"` // 允许空密码
}

// OutputConfig 输出配置
type OutputConfig struct {
	Format     string `yaml:"format"`      // json, text, csv
	File       string `yaml:"file"`        // 输出文件
	Verbose    bool   `yaml:"verbose"`     // 详细输出
	Silent     bool   `yaml:"silent"`      // 静默模式
	NoColor    bool   `yaml:"no_color"`    // 禁用颜色
	ShowFailed bool   `yaml:"show_failed"` // 显示失败结果
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Type     string `yaml:"type"`     // http, https, socks5
	Address  string `yaml:"address"`  // 代理地址
	Username string `yaml:"username"` // 代理用户名
	Password string `yaml:"password"` // 代理密码
}

// DefaultConfig 返回默认配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Version:  "1.0.0",
		Debug:    false,
		LogLevel: "info",
		Brute: BruteConfig{
			TargetConcurrent:   50,
			TaskConcurrent:     1,
			MinDelay:           "100ms",
			MaxDelay:           "500ms",
			Timeout:            "10s",
			MaxRetries:         3,
			OkToStop:           false,
			FinishingThreshold: 0,
			SkipEmptyPassword:  true,
			SkipEmptyUsername:  true,
			OnlyNeedPassword:   false,
			DefaultUserDict: []string{
				"admin", "root", "administrator", "user", "test", "guest",
			},
			DefaultPassDict: []string{
				"123456", "password", "admin", "root", "123456789", "12345678",
			},
		},
		Output: OutputConfig{
			Format:     "text",
			Verbose:    false,
			Silent:     false,
			NoColor:    false,
			ShowFailed: false,
		},
		Proxy: ProxyConfig{
			Enabled: false,
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*AppConfig, error) {
	config := DefaultConfig()

	if filename == "" {
		return config, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *AppConfig, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ParseDuration 解析持续时间字符串
func (c *BruteConfig) ParseMinDelay() (time.Duration, error) {
	return time.ParseDuration(c.MinDelay)
}

// ParseMaxDelay 解析最大延迟
func (c *BruteConfig) ParseMaxDelay() (time.Duration, error) {
	return time.ParseDuration(c.MaxDelay)
}

// ParseTimeout 解析超时时间
func (c *BruteConfig) ParseTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Timeout)
}

// Validate 验证配置
func (c *AppConfig) Validate() error {
	// 验证爆破配置
	if c.Brute.TargetConcurrent <= 0 {
		return fmt.Errorf("target_concurrent must be positive")
	}

	if c.Brute.TaskConcurrent <= 0 {
		return fmt.Errorf("task_concurrent must be positive")
	}

	// 验证时间配置
	if _, err := c.Brute.ParseMinDelay(); err != nil {
		return fmt.Errorf("invalid min_delay: %w", err)
	}

	if _, err := c.Brute.ParseMaxDelay(); err != nil {
		return fmt.Errorf("invalid max_delay: %w", err)
	}

	if _, err := c.Brute.ParseTimeout(); err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	// 验证输出格式
	validFormats := map[string]bool{
		"json": true,
		"text": true,
		"csv":  true,
	}

	if !validFormats[c.Output.Format] {
		return fmt.Errorf("invalid output format: %s", c.Output.Format)
	}

	// 验证代理配置
	if c.Proxy.Enabled {
		if c.Proxy.Address == "" {
			return fmt.Errorf("proxy address cannot be empty when proxy is enabled")
		}

		validTypes := map[string]bool{
			"http":   true,
			"https":  true,
			"socks5": true,
		}

		if !validTypes[c.Proxy.Type] {
			return fmt.Errorf("invalid proxy type: %s", c.Proxy.Type)
		}
	}

	return nil
}
