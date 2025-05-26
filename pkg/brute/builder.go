package brute

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
)

// protocolHandlers 协议处理器映射
var protocolHandlers = map[string]BruteCallback{}

// GetProtocolHandler 获取协议处理器
func GetProtocolHandler(protocol string) (BruteCallback, bool) {
	handler, exists := protocolHandlers[protocol]
	return handler, exists
}

// RegisterProtocolHandler 注册协议处理器
func RegisterProtocolHandler(protocol string, handler BruteCallback) {
	protocolHandlers[protocol] = handler
}

// GetSupportedProtocols 获取支持的协议列表
func GetSupportedProtocols() []string {
	return lo.Keys(protocolHandlers)
}

// Builder 爆破引擎构建器
type Builder struct {
	config   *Config
	targets  []Target
	userDict []string
	passDict []string
	callback ResultCallback
	ctx      context.Context
}

// Target 目标定义
type Target struct {
	Type string `json:"type"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// NewBuilder 创建新的构建器
func NewBuilder(ctx context.Context) *Builder {
	return &Builder{
		config:   DefaultConfig(),
		targets:  make([]Target, 0),
		userDict: make([]string, 0),
		passDict: make([]string, 0),
		ctx:      ctx,
	}
}

// WithConfig 设置配置
func (b *Builder) WithConfig(config *Config) *Builder {
	b.config = config
	return b
}

// WithTargets 设置目标列表
func (b *Builder) WithTargets(targets []Target) *Builder {
	b.targets = targets
	return b
}

// WithTarget 添加单个目标
func (b *Builder) WithTarget(serviceType, host string, port int) *Builder {
	b.targets = append(b.targets, Target{
		Type: serviceType,
		Host: host,
		Port: port,
	})
	return b
}

// WithUserDict 设置用户字典
func (b *Builder) WithUserDict(users []string) *Builder {
	b.userDict = users
	return b
}

// WithPassDict 设置密码字典
func (b *Builder) WithPassDict(passwords []string) *Builder {
	b.passDict = passwords
	return b
}

// WithUserDictFile 设置用户字典文件
func (b *Builder) WithUserDictFile(filename string) *Builder {
	b.config.UserDictFile = filename
	return b
}

// WithPassDictFile 设置密码字典文件
func (b *Builder) WithPassDictFile(filename string) *Builder {
	b.config.PassDictFile = filename
	return b
}

// WithResultCallback 设置结果回调
func (b *Builder) WithResultCallback(callback ResultCallback) *Builder {
	b.callback = callback
	return b
}

// WithConcurrency 设置并发数
func (b *Builder) WithConcurrency(targetConcurrent, taskConcurrent int) *Builder {
	b.config.TargetConcurrent = targetConcurrent
	b.config.TaskConcurrent = taskConcurrent
	return b
}

// WithTimeout 设置超时
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.config.Timeout = timeout
	return b
}

// WithDelay 设置延迟
func (b *Builder) WithDelay(min, max time.Duration) *Builder {
	b.config.MinDelay = min
	b.config.MaxDelay = max
	return b
}

// WithRetries 设置重试次数
func (b *Builder) WithRetries(retries int) *Builder {
	b.config.MaxRetries = retries
	return b
}

// WithOkToStop 设置成功后停止
func (b *Builder) WithOkToStop(okToStop bool) *Builder {
	b.config.OkToStop = okToStop
	return b
}

// WithFinishingThreshold 设置完成阈值
func (b *Builder) WithFinishingThreshold(threshold int) *Builder {
	b.config.FinishingThreshold = threshold
	return b
}

// WithCustomCallback 设置自定义回调
func (b *Builder) WithCustomCallback(callback BruteCallback) *Builder {
	b.config.CustomCallback = callback
	return b
}

// Build 构建爆破引擎
func (b *Builder) Build() (*Engine, error) {
	// 合并字典
	b.config.UserDict = append(b.config.UserDict, b.userDict...)
	b.config.PassDict = append(b.config.PassDict, b.passDict...)

	// 创建引擎
	engine, err := NewEngine(b.ctx, b.config)
	if err != nil {
		return nil, err
	}

	// 设置结果回调
	if b.callback != nil {
		engine.SetResultCallback(b.callback)
	}

	// 添加目标
	for _, target := range b.targets {
		engine.AddTarget(target.Type, target.Host, target.Port)
	}

	// 生成爆破任务
	if err := b.generateBruteItems(engine); err != nil {
		return nil, err
	}

	return engine, nil
}

// generateBruteItems 生成爆破任务
func (b *Builder) generateBruteItems(engine *Engine) error {
	for _, target := range b.targets {
		for _, username := range b.config.UserDict {
			// 跳过空用户名
			if b.config.SkipEmptyUsername && username == "" {
				continue
			}

			for _, password := range b.config.PassDict {
				// 跳过空密码
				if b.config.SkipEmptyPassword && password == "" {
					continue
				}

				item := &BruteItem{
					Type:     target.Type,
					Target:   target.Host,
					Port:     target.Port,
					Username: username,
					Password: password,
					Context:  b.ctx,
					Timeout:  b.config.Timeout,
					Extra:    make(map[string]string),
				}
				if err := engine.Feed(item); err != nil {
					return fmt.Errorf("failed to feed brute item: %w", err)
				}
			}
		}
	}

	return nil
}

// QuickBrute 快速爆破函数
func QuickBrute(ctx context.Context, protocol, host string, port int, users, passwords []string, callback ResultCallback) error {
	builder := NewBuilder(ctx).
		WithTarget(protocol, host, port).
		WithUserDict(users).
		WithPassDict(passwords).
		WithResultCallback(callback)

	engine, err := builder.Build()
	if err != nil {
		return err
	}

	return engine.Start()
}

// BatchBrute 批量爆破函数
func BatchBrute(ctx context.Context, targets []Target, users, passwords []string, callback ResultCallback) error {
	return BatchBruteWithConfig(ctx, targets, users, passwords, callback, nil)
}

// BatchBruteWithConfig 带配置的批量爆破函数
func BatchBruteWithConfig(ctx context.Context, targets []Target, users, passwords []string, callback ResultCallback, config *Config) error {
	builder := NewBuilder(ctx).
		WithTargets(targets).
		WithUserDict(users).
		WithPassDict(passwords).
		WithResultCallback(callback)

	// 如果提供了配置，使用配置
	if config != nil {
		builder = builder.WithConfig(config)
	}

	engine, err := builder.Build()
	if err != nil {
		return err
	}

	return engine.Start()
}
