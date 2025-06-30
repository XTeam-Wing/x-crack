package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	_ "github.com/XTeam-Wing/x-crack/pkg/protocols" // 导入协议包以注册处理器
	"github.com/XTeam-Wing/x-crack/pkg/utils"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	fileutil "github.com/projectdiscovery/utils/file"
	"github.com/samber/lo"

	folderutil "github.com/projectdiscovery/utils/folder"
)

var (
	defaultConfigLocation = filepath.Join(folderutil.AppConfigDirOrDefault(".config", "x-crack"), "config.yaml")
)

// CLI 命令行参数
type CLI struct {
	// 目标设置
	Target        string              `json:"target"`         // 目标地址
	Targets       goflags.StringSlice `json:"targets"`        // 目标列表
	TargetFile    string              `json:"target_file"`    // 目标文件
	ServiceTarget string              `json:"service_target"` // 服务目标文件 (protocol://host:port格式)
	Port          int                 `json:"port"`           // 单个端口
	Ports         string              `json:"ports"`          // 端口范围
	PortFile      string              `json:"port_file"`      // 端口文件
	Protocol      string              `json:"protocol"`       // 协议类型
	Protocols     goflags.StringSlice `json:"protocols"`      // 协议列表

	// 认证设置
	Username     string              `json:"username"`      // 单个用户名
	Usernames    goflags.StringSlice `json:"usernames"`     // 用户名列表
	UserFile     string              `json:"user_file"`     // 用户名文件
	Password     string              `json:"password"`      // 单个密码
	Passwords    goflags.StringSlice `json:"passwords"`     // 密码列表
	PassFile     string              `json:"pass_file"`     // 密码文件
	UserPassFile string              `json:"userpass_file"` // 用户名:密码文件

	// 爆破设置
	TargetConcurrent int    `json:"target_concurrent"` // 目标并发数
	TaskConcurrent   int    `json:"task_concurrent"`   // 任务并发数
	Delay            string `json:"delay"`             // 延迟
	Timeout          string `json:"timeout"`           // 超时
	Retries          int    `json:"retries"`           // 重试次数
	OkToStop         bool   `json:"ok_to_stop"`        // 成功后停止

	// 空凭据设置
	AllowBlankUsername bool `json:"allow_blank_username"` // 允许空用户名
	AllowBlankPassword bool `json:"allow_blank_password"` // 允许空密码

	// 输出设置
	Output       string `json:"output"`        // 输出文件
	Format       string `json:"format"`        // 输出格式
	Verbose      bool   `json:"verbose"`       // 详细输出
	Debug        bool   `json:"debug"`         // 调试模式
	Silent       bool   `json:"silent"`        // 静默模式
	NoColor      bool   `json:"no_color"`      // 禁用颜色
	ShowFailed   bool   `json:"show_failed"`   // 显示失败结果
	ShowProgress bool   `json:"show_progress"` // 显示进度条

	// 其他设置
	ConfigFile string `json:"config_file"` // 配置文件
	Version    bool   `json:"version"`     // 显示版本
}

var (
	successCount int32
	failureCount int32
	totalCount   int32
)

func main() {
	// 解析命令行参数
	cli, err := parseFlags()
	if err != nil {
		gologger.Fatal().Msgf("Failed to parse flags: %v", err)
	}

	// 显示版本信息
	if cli.Version {
		fmt.Println("x-crack version 1.0.0")
		return
	}

	// 设置日志级别
	if cli.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
		gologger.Debug().Msg("Verbose mode enabled")
	} else if cli.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	} else if cli.Debug {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelDebug)
		gologger.Debug().Msg("Debug mode enabled")
	} else {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelInfo)
	}

	// 验证参数
	if err := validateCLI(cli); err != nil {
		gologger.Fatal().Msgf("Invalid parameters: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	setupSignalHandler(cancel)

	// 执行爆破
	if err := executeBrute(ctx, cli); err != nil {
		gologger.Fatal().Msgf("Brute force failed: %v", err)
	}

	// 显示统计信息
	showStatistics()
}

// executeBrute 执行爆破
func executeBrute(ctx context.Context, cli *CLI) error {
	// 检查是否使用服务目标文件
	if cli.ServiceTarget != "" {
		return executeBruteWithServiceTargets(ctx, cli)
	}

	// 解析目标
	targets, err := parseTargets(cli)
	if err != nil {
		return fmt.Errorf("failed to parse targets: %w", err)
	}

	// 解析协议
	protocols, err := parseProtocols(cli)
	if err != nil {
		return fmt.Errorf("failed to parse protocols: %w", err)
	}

	// 解析端口
	ports, err := parsePorts(cli)
	if err != nil {
		return fmt.Errorf("failed to parse ports: %w", err)
	}

	// 解析用户名和密码
	usernames, passwords, err := parseCredentials(cli)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}
	// 创建爆破配置
	bruteConfig := createBruteConfig(cli)

	if cli.AllowBlankPassword {
		bruteConfig.AllowBlankUsername = true
	}
	if cli.AllowBlankUsername {
		bruteConfig.AllowBlankUsername = true
	}
	// 创建结果回调
	resultCallback := createResultCallback(cli)

	// 构建目标列表
	var bruteTargets []brute.Target
	for _, target := range targets {
		for _, protocol := range protocols {
			targetPorts := ports
			if len(targetPorts) == 0 {
				targetPorts = utils.GetDefaultPorts(protocol)
			}

			for _, port := range targetPorts {
				bruteTargets = append(bruteTargets, brute.Target{
					Type: protocol,
					Host: target,
					Port: port,
				})
			}
		}
	}

	atomic.AddInt32(&totalCount, int32(len(bruteTargets)*len(usernames)*len(passwords)))
	// 执行批量爆破
	return brute.BatchBruteWithConfig(ctx, bruteTargets, usernames, passwords, resultCallback, bruteConfig)
}

// executeBruteWithServiceTargets 使用服务目标文件执行爆破
func executeBruteWithServiceTargets(ctx context.Context, cli *CLI) error {
	// 解析服务目标文件
	serviceTargets, err := utils.ParseServiceTargetFile(cli.ServiceTarget)
	if err != nil {
		return fmt.Errorf("failed to parse service target file: %w", err)
	}

	gologger.Info().Msgf("Loaded %d service targets from file: %s", len(serviceTargets), cli.ServiceTarget)

	// 解析用户名和密码
	usernames, passwords, err := parseCredentials(cli)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	// 创建爆破配置
	bruteConfig := createBruteConfig(cli)

	if cli.AllowBlankPassword {
		bruteConfig.AllowBlankPassword = true
	}
	if cli.AllowBlankUsername {
		bruteConfig.AllowBlankUsername = true
	}

	// 创建结果回调
	resultCallback := createResultCallback(cli)

	// 转换服务目标为爆破目标
	var bruteTargets []brute.Target
	for _, serviceTarget := range serviceTargets {
		// 验证服务目标
		if err := utils.ValidateServiceTarget(serviceTarget); err != nil {
			gologger.Warning().Msgf("Skipping invalid service target %s://%s:%d: %v",
				serviceTarget.Protocol, serviceTarget.Host, serviceTarget.Port, err)
			continue
		}

		bruteTargets = append(bruteTargets, brute.Target{
			Type: serviceTarget.Protocol,
			Host: serviceTarget.Host,
			Port: serviceTarget.Port,
		})
	}

	if len(bruteTargets) == 0 {
		return fmt.Errorf("no valid brute targets found from service target file")
	}

	gologger.Info().Msgf("Starting brute force on %d targets with %d usernames and %d passwords",
		len(bruteTargets), len(usernames), len(passwords))

	atomic.AddInt32(&totalCount, int32(len(bruteTargets)*len(usernames)*len(passwords)))

	// 执行批量爆破
	return brute.BatchBruteWithConfig(ctx, bruteTargets, usernames, passwords, resultCallback, bruteConfig)
}

// parseTargets 解析目标
func parseTargets(cli *CLI) ([]string, error) {
	var targets []string

	// 单个目标
	if cli.Target != "" {
		targets = append(targets, cli.Target)
	}

	// 目标列表
	targets = append(targets, []string(cli.Targets)...)

	// 目标文件
	if cli.TargetFile != "" {
		fileTargets, err := utils.LoadLinesFromFile(cli.TargetFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load targets from file: %w", err)
		}
		targets = append(targets, fileTargets...)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets specified")
	}

	// 解析和验证IP地址和子网
	var validTargets []string
	for _, target := range targets {
		validTargets = append(validTargets, utils.ParseIP(target)...)
	}

	return lo.Uniq(validTargets), nil
}

// parseProtocols 解析协议
func parseProtocols(cli *CLI) ([]string, error) {
	var protocols []string

	// 单个协议
	if cli.Protocol != "" {
		protocols = append(protocols, cli.Protocol)
	}

	// 协议列表
	protocols = append(protocols, []string(cli.Protocols)...)

	if len(protocols) == 0 {
		return nil, fmt.Errorf("no protocols specified")
	}

	// 验证协议
	supportedProtocols := brute.GetSupportedProtocols()
	for _, protocol := range protocols {
		if !lo.Contains(supportedProtocols, protocol) {
			return nil, fmt.Errorf("unsupported protocol: %s", protocol)
		}
	}

	return lo.Uniq(protocols), nil
}

// parsePorts 解析端口
func parsePorts(cli *CLI) ([]int, error) {
	var ports []int

	// 单个端口
	if cli.Port > 0 {
		ports = append(ports, cli.Port)
	}

	// 端口范围
	if cli.Ports != "" {
		rangePorts, err := utils.ParsePortRange(cli.Ports)
		if err != nil {
			return nil, fmt.Errorf("failed to parse port range: %w", err)
		}
		ports = append(ports, rangePorts...)
	}

	// 端口文件
	if cli.PortFile != "" {
		lines, err := utils.LoadLinesFromFile(cli.PortFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load ports from file: %w", err)
		}

		for _, line := range lines {
			rangePorts, err := utils.ParsePortRange(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse port line %s: %w", line, err)
			}
			ports = append(ports, rangePorts...)
		}
	}

	return lo.Uniq(ports), nil
}

// parseCredentials 解析认证信息
func parseCredentials(cli *CLI) ([]string, []string, error) {
	var usernames, passwords []string

	// 解析用户名
	if cli.Username != "" {
		usernames = append(usernames, cli.Username)
	}
	usernames = append(usernames, []string(cli.Usernames)...)

	if cli.UserFile != "" {
		fileUsers, err := utils.LoadLinesFromFile(cli.UserFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load usernames: %w", err)
		}
		usernames = append(usernames, fileUsers...)
	}

	// 解析密码
	if cli.Password != "" {
		passwords = append(passwords, cli.Password)
	}
	passwords = append(passwords, []string(cli.Passwords)...)

	if cli.PassFile != "" {
		filePasswords, err := utils.LoadLinesFromFile(cli.PassFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load passwords: %w", err)
		}
		passwords = append(passwords, filePasswords...)
	}

	// 解析用户名:密码文件
	if cli.UserPassFile != "" {
		lines, err := utils.LoadLinesFromFile(cli.UserPassFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load user:pass file: %w", err)
		}

		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				usernames = append(usernames, parts[0])
				passwords = append(passwords, parts[1])
			}
		}
	}
	// 添加空凭据支持
	if cli.AllowBlankUsername {
		usernames = append(usernames, "")
	}
	if cli.AllowBlankPassword {
		passwords = append(passwords, "")
	}
	return lo.Uniq(usernames), lo.Uniq(passwords), nil
}

// createBruteConfig 创建爆破配置
func createBruteConfig(cli *CLI) *brute.Config {
	config := brute.DefaultConfig()
	// 设置进度
	if cli.ShowProgress {
		config.ShowProgress = true
	}
	// 设置并发数
	if cli.TargetConcurrent > 0 {
		config.TargetConcurrent = cli.TargetConcurrent
	}
	// 如果都没有设置，保持默认值

	if cli.TaskConcurrent > 0 {
		config.TaskConcurrent = cli.TaskConcurrent
	}
	// 如果都没有设置，保持默认值

	// 设置延迟
	if cli.Delay != "" {
		if delay, err := time.ParseDuration(cli.Delay); err == nil {
			config.MinDelay = delay
			config.MaxDelay = delay
		}
	}

	// 设置超时
	if cli.Timeout != "" {
		if timeout, err := time.ParseDuration(cli.Timeout); err == nil {
			config.Timeout = timeout
		}
	}

	// 设置重试
	if cli.Retries > 0 {
		config.MaxRetries = cli.Retries
	}
	// 如果都没有设置，保持默认值

	// 设置停止条件
	config.OkToStop = cli.OkToStop

	// 设置跳过空值选项
	// 如果用户明确允许空凭据，则不跳过它们
	if cli.AllowBlankUsername {
		config.SkipEmptyUsername = false
	}

	if cli.AllowBlankPassword {
		config.SkipEmptyPassword = false
	}
	return config
}

// createResultCallback 创建结果回调
func createResultCallback(cli *CLI) brute.ResultCallback {
	var outputFile *os.File
	if cli.Output != "" {
		file, err := os.Create(cli.Output)
		if err != nil {
			gologger.Error().Msgf("Failed to create output file: %v", err)
		} else {
			outputFile = file
		}
	}

	return func(result *brute.BruteResult) {
		if result.Success {
			atomic.AddInt32(&successCount, 1)

			// 输出成功结果
			switch cli.Format {
			case "json":
				data, _ := json.Marshal(result)
				gologger.Info().Msgf("成功结果: %s", string(data))
			default:
				gologger.Info().Msgf("%s", result.String())
			}

			// 写入文件
			if outputFile != nil {
				switch cli.Format {
				case "json":
					data, _ := json.Marshal(result)
					outputFile.WriteString(string(data) + "\n")
				default:
					outputFile.WriteString(fmt.Sprintf("[SUCCESS] %s\n", result.String()))
				}
			}
		} else {
			atomic.AddInt32(&failureCount, 1)

			// 显示失败结果
			if cli.ShowFailed {
				switch cli.Format {
				case "json":
					data, _ := json.Marshal(result)
					fmt.Println(string(data))
				default:
					fmt.Printf("[FAILED] %s\n", result.String())
				}
			}
		}
	}
}

// parseFlags 解析命令行参数（简化实现，实际应该使用flag或其他库）
func parseFlags() (*CLI, error) {
	cli := &CLI{
		TargetConcurrent: 0,  // 使用0作为默认值，表示未设置
		TaskConcurrent:   0,  // 使用0作为默认值，表示未设置
		Timeout:          "", // 使用空字符串作为默认值，表示未设置
		Retries:          0,  // 使用0作为默认值，表示未设置
		Format:           "text",
	}

	flagSet := goflags.NewFlagSet()
	flagSet.CreateGroup("target", "Target settings",
		flagSet.StringVar(&cli.Target, "target", "", "Target host (e.g. 192.168.1.1)"),
		flagSet.StringSliceVar(&cli.Targets, "targets", []string{}, "Target hosts (comma separated) (eg. 192.168.1.1/24,192.168.1.1-3)", goflags.NormalizedStringSliceOptions),
		flagSet.StringVarP(&cli.TargetFile, "target-file", "l", "", "File containing target hosts (eg. 192.168.1.1/24,192.168.1.1-3)"),
		flagSet.StringVar(&cli.ServiceTarget, "service-target", "", "File containing service targets in protocol://host:port format (e.g. telnet://1.1.1.1:23)"),
		flagSet.IntVar(&cli.Port, "port", 0, "Target port"),
		flagSet.StringVar(&cli.Ports, "ports", "", "Port range (e.g. 22,3389,1433-1434)"),
		flagSet.StringVar(&cli.PortFile, "port-file", "", "File containing ports"),
		flagSet.StringVar(&cli.Protocol, "protocol", "", "Protocol to use (ssh,mysql,ftp,etc.)"),
		flagSet.StringSliceVar(&cli.Protocols, "protocols", []string{}, "Protocols to use (comma separated)", goflags.NormalizedStringSliceOptions),
	)

	flagSet.CreateGroup("auth", "Authentication settings",
		flagSet.StringVarP(&cli.Username, "username", "u", "", "Username for authentication"),
		flagSet.StringSliceVar(&cli.Usernames, "usernames", []string{}, "Usernames (comma separated)", goflags.NormalizedStringSliceOptions),
		flagSet.StringVarP(&cli.UserFile, "user-file", "uf", "", "File containing usernames"),
		flagSet.StringVarP(&cli.Password, "password", "p", "", "Password for authentication"),
		flagSet.StringSliceVar(&cli.Passwords, "passwords", []string{}, "Passwords (comma separated)", goflags.NormalizedStringSliceOptions),
		flagSet.StringVarP(&cli.PassFile, "pass-file", "pf", "", "File containing passwords"),
		flagSet.StringVar(&cli.UserPassFile, "userpass-file", "", "File containing username:password combinations"),
		flagSet.BoolVar(&cli.AllowBlankUsername, "allow-blank-username", false, "Allow blank/empty usernames during brute force"),
		flagSet.BoolVar(&cli.AllowBlankPassword, "allow-blank-password", false, "Allow blank/empty passwords during brute force"),
	)

	flagSet.CreateGroup("brute", "Brute force settings",
		flagSet.IntVar(&cli.TargetConcurrent, "target-concurrent", 10, "Number of concurrent targets"),
		flagSet.IntVar(&cli.TaskConcurrent, "task-concurrent", 10, "Number of concurrent tasks per target"),
		flagSet.StringVar(&cli.Delay, "delay", "", "Delay between requests (e.g. 100ms)"),
		flagSet.StringVar(&cli.Timeout, "timeout", "10s", "Timeout for each request"),
		flagSet.IntVar(&cli.Retries, "retries", 3, "Number of retries for failed requests"),
		flagSet.BoolVarP(&cli.OkToStop, "ok-to-stop", "ots", false, "Stop after first successful authentication"),
	)

	flagSet.CreateGroup("output", "Output settings",
		flagSet.StringVar(&cli.Output, "output", "", "Output file path"),
		flagSet.StringVar(&cli.Format, "format", "text", "Output format (text,json,csv)"),
		flagSet.BoolVarP(&cli.Verbose, "verbose", "v", false, "Verbose output"),
		flagSet.BoolVarP(&cli.Debug, "debug", "d", false, "Debug mode"),
		flagSet.BoolVar(&cli.Silent, "silent", false, "Silent mode"),
		flagSet.BoolVar(&cli.NoColor, "no-color", false, "Disable colored output"),
		flagSet.BoolVar(&cli.ShowFailed, "show-failed", false, "Show failed authentication attempts"),
		flagSet.BoolVarP(&cli.ShowProgress, "show-progress", "sp", false, "Show progress bar during brute force"),
	)

	flagSet.CreateGroup("misc", "Miscellaneous settings",
		flagSet.StringVar(&cli.ConfigFile, "config", defaultConfigLocation, "Configuration file path"),
		flagSet.BoolVar(&cli.Version, "version", false, "Show version information"),
	)
	// 其他设置
	if err := flagSet.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}
	if cli.ConfigFile != defaultConfigLocation {
		_ = cli.loadConfigFrom(cli.ConfigFile)
	}
	return cli, nil
}

func (c *CLI) loadConfigFrom(location string) error {
	if !fileutil.FileExists(location) {
		return fmt.Errorf("config file %s does not exist", location)
	}
	return fileutil.Unmarshal(fileutil.YAML, []byte(location), c)
}

// validateCLI 验证命令行参数
func validateCLI(cli *CLI) error {
	if cli.Target == "" && len(cli.Targets) == 0 && cli.TargetFile == "" && cli.ServiceTarget == "" {
		return fmt.Errorf("no targets specified")
	}

	// 如果使用了service-target，则不需要检查协议参数，因为协议信息已包含在服务URL中
	if cli.ServiceTarget == "" && cli.Protocol == "" && len(cli.Protocols) == 0 {
		return fmt.Errorf("no protocols specified")
	}

	return nil
}

// setupSignalHandler 设置信号处理
func setupSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		gologger.Info().Msg("Received interrupt signal, stopping...")
		cancel()
	}()
}

// showStatistics 显示统计信息
func showStatistics() {
	total := atomic.LoadInt32(&totalCount)
	success := atomic.LoadInt32(&successCount)
	failure := atomic.LoadInt32(&failureCount)

	// gologger.Info().Msgf("Brute force completed")
	gologger.Info().Msgf("Total: %d, Success: %d, Failed: %d", total, success, failure)
}
