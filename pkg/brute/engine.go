package brute

import (
	"bufio"
	"container/list"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/projectdiscovery/gologger"
	"github.com/samber/lo"
	"golang.org/x/time/rate"
)

// Engine 爆破引擎
type Engine struct {
	config         *Config
	targets        *list.List
	processes      sync.Map
	resultCallback ResultCallback
	limiter        *rate.Limiter
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	targetWg       sync.WaitGroup
	globalSem      chan struct{} // 全局并发控制信号量
}

// targetProcess 目标处理状态
type targetProcess struct {
	Target    string
	Items     []*BruteItem
	Count     int32
	Finished  bool
	mutex     sync.RWMutex
	semaphore chan struct{}
}

// NewEngine 创建新的爆破引擎
func NewEngine(ctx context.Context, config *Config) (*Engine, error) {
	if config == nil {
		config = DefaultConfig()
	}
	// protocols.RegisterAllProtocols()

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 加载字典
	if err := loadDictionaries(config); err != nil {
		return nil, fmt.Errorf("failed to load dictionaries: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	// 创建限流器 - 修复：使用正确的速率参数
	// rate.Every(config.MinDelay) 表示每隔 MinDelay 时间允许一个请求
	// config.TargetConcurrent 表示突发容量（桶的大小）
	limiter := rate.NewLimiter(rate.Every(config.MinDelay), config.TargetConcurrent)
	engine := &Engine{
		config:    config,
		targets:   list.New(),
		limiter:   limiter,
		ctx:       ctx,
		cancel:    cancel,
		globalSem: make(chan struct{}, config.TargetConcurrent), // 全局并发控制
	}

	return engine, nil
}

// SetResultCallback 设置结果回调
func (e *Engine) SetResultCallback(callback ResultCallback) {
	e.resultCallback = callback
}

// AddTarget 添加目标
func (e *Engine) AddTarget(serviceType, target string, port int) {
	targetKey := fmt.Sprintf("%s:%s:%d", serviceType, target, port)
	e.targets.PushBack(targetKey)

	// 初始化目标处理器
	process := &targetProcess{
		Target:    targetKey,
		Items:     make([]*BruteItem, 0),
		semaphore: make(chan struct{}, e.config.TaskConcurrent),
	}
	e.processes.Store(targetKey, process)
}

// Feed 向引擎提供爆破任务
func (e *Engine) Feed(item *BruteItem) error {
	targetKey := fmt.Sprintf("%s:%s:%d", item.Type, item.Target, item.Port)

	processRaw, ok := e.processes.Load(targetKey)
	if !ok {
		return fmt.Errorf("target %s not found", targetKey)
	}

	process := processRaw.(*targetProcess)
	process.mutex.Lock()
	process.Items = append(process.Items, item)
	process.mutex.Unlock()

	return nil
}

// Start 开始爆破
func (e *Engine) Start() error {
	gologger.Info().Msg("Starting brute force engine")
	gologger.Info().Msgf("Configuration: TargetConcurrent=%d, TaskConcurrent=%d, MinDelay=%v",
		e.config.TargetConcurrent, e.config.TaskConcurrent, e.config.MinDelay)

	targetCount := e.targets.Len()
	if targetCount == 0 {
		gologger.Warning().Msg("No targets to process")
		return nil
	}

	gologger.Info().Msgf("Processing %d targets", targetCount)

	// 遍历所有目标
	for element := e.targets.Front(); element != nil; element = element.Next() {
		targetKey := element.Value.(string)

		e.targetWg.Add(1)
		go e.processTarget(targetKey)
	}

	// 等待所有目标处理完成
	e.targetWg.Wait()
	processedCount := e.GetProcessedCount()
	gologger.Info().Msgf("Brute force engine completed, processed %d items", processedCount)

	return nil
}

// Stop 停止爆破
func (e *Engine) Stop() {
	gologger.Info().Msg("Stopping brute force engine")
	e.cancel()
	e.wg.Wait()
}

// processTarget 处理单个目标
func (e *Engine) processTarget(targetKey string) {
	defer e.targetWg.Done()

	processRaw, ok := e.processes.Load(targetKey)
	if !ok {
		gologger.Error().Msgf("Target process not found: %s", targetKey)
		return
	}

	process := processRaw.(*targetProcess)
	var itemWg sync.WaitGroup

	// 处理所有任务项
	for _, item := range process.Items {
		// 检查上下文
		select {
		case <-e.ctx.Done():
			return
		default:
		}

		// 检查是否需要提前停止
		process.mutex.RLock()
		finished := process.Finished
		process.mutex.RUnlock()
		if finished {
			break
		}

		// 获取全局信号量，控制整体并发数
		select {
		case e.globalSem <- struct{}{}:
			// 然后获取目标级别的信号量，控制单个目标的并发数
			select {
			case process.semaphore <- struct{}{}:
				itemWg.Add(1)
				e.wg.Add(1)
				gologger.Debug().Msgf("Processing target: %s service: %s username:%s password:%s",
					targetKey, item.Type, item.Username, item.Password)
				go e.processItem(item, process, &itemWg)
			case <-e.ctx.Done():
				<-e.globalSem // 释放全局信号量
				return
			}
		case <-e.ctx.Done():
			return
		}
	}

	// 等待当前目标的所有任务完成
	itemWg.Wait()
	gologger.Debug().Msgf("Target %s processing completed", targetKey)
}

// processItem 处理单个爆破项
func (e *Engine) processItem(item *BruteItem, process *targetProcess, itemWg *sync.WaitGroup) {
	defer e.wg.Done()
	defer itemWg.Done()
	defer func() {
		<-process.semaphore // 释放目标级别信号量
		<-e.globalSem       // 释放全局信号量
	}()

	// 限流 - 等待限流器允许
	if err := e.limiter.Wait(e.ctx); err != nil {
		gologger.Debug().Msgf("Rate limiter wait failed: %v", err)
		return
	}

	// 执行爆破
	startTime := time.Now()
	result := e.executeItem(item)
	result.ResponseTime = time.Since(startTime)

	// 更新计数
	atomic.AddInt32(&process.Count, 1)

	// 调用结果回调
	if e.resultCallback != nil {
		e.resultCallback(result)
	}

	// 如果成功且配置为成功后停止，则停止处理
	if result.Success && e.config.OkToStop {
		process.mutex.Lock()
		process.Finished = true
		process.mutex.Unlock()
		gologger.Info().Msgf("Success found for target %s, stopping further attempts", process.Target)
		return
	}
}

// executeItem 执行单个爆破项
func (e *Engine) executeItem(item *BruteItem) *BruteResult {
	result := &BruteResult{
		Item:    item,
		Success: false,
	}

	// 如果有自定义回调，使用自定义回调
	if e.config.CustomCallback != nil {
		return e.config.CustomCallback(item)
	}

	// 否则使用内置的协议处理器
	handler, exists := GetProtocolHandler(item.Type)
	if !exists {
		gologger.Error().Msgf("Unsupported protocol: %s", item.Type)
		result.Error = fmt.Errorf("unsupported protocol: %s", item.Type)
		return result
	}
	return handler(item)
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.TargetConcurrent <= 0 {
		return fmt.Errorf("target concurrent must be positive, got: %d", config.TargetConcurrent)
	}
	if config.TaskConcurrent <= 0 {
		return fmt.Errorf("task concurrent must be positive, got: %d", config.TaskConcurrent)
	}
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %v", config.Timeout)
	}
	if config.MinDelay < 0 {
		return fmt.Errorf("min delay cannot be negative, got: %v", config.MinDelay)
	}
	if config.MaxDelay > 0 && config.MinDelay > config.MaxDelay {
		return fmt.Errorf("min delay (%v) cannot be greater than max delay (%v)", config.MinDelay, config.MaxDelay)
	}

	// 合理性检查
	if config.TargetConcurrent > 1000 {
		gologger.Warning().Msgf("High target concurrent value: %d, this may cause performance issues", config.TargetConcurrent)
	}
	if config.TaskConcurrent > 100 {
		gologger.Warning().Msgf("High task concurrent value: %d, this may cause performance issues", config.TaskConcurrent)
	}

	return nil
}

// loadDictionaries 加载字典
func loadDictionaries(config *Config) error {
	// 加载用户字典文件
	if config.UserDictFile != "" {
		userDict, err := loadDictFromFile(config.UserDictFile)
		if err != nil {
			return fmt.Errorf("failed to load user dictionary: %w", err)
		}
		config.UserDict = append(config.UserDict, userDict...)
	}

	// 加载密码字典文件
	if config.PassDictFile != "" {
		passDict, err := loadDictFromFile(config.PassDictFile)
		if err != nil {
			return fmt.Errorf("failed to load password dictionary: %w", err)
		}
		config.PassDict = append(config.PassDict, passDict...)
	}

	// 去重
	config.UserDict = lo.Uniq(config.UserDict)
	config.PassDict = lo.Uniq(config.PassDict)

	return nil
}

// loadDictFromFile 从文件加载字典
func loadDictFromFile(filename string) ([]string, error) {
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

// GetTargetCount 获取目标数量
func (e *Engine) GetTargetCount() int {
	return e.targets.Len()
}

// GetProcessedCount 获取已处理数量
func (e *Engine) GetProcessedCount() int32 {
	var total int32
	e.processes.Range(func(key, value interface{}) bool {
		process := value.(*targetProcess)
		total += atomic.LoadInt32(&process.Count)
		return true
	})
	return total
}

// UpdateRateLimit 动态更新限流器设置
func (e *Engine) UpdateRateLimit(minDelay time.Duration, burstSize int) {
	if minDelay <= 0 || burstSize <= 0 {
		gologger.Warning().Msg("Invalid rate limit parameters, ignoring update")
		return
	}

	// 创建新的限流器
	newLimiter := rate.NewLimiter(rate.Every(minDelay), burstSize)
	e.limiter = newLimiter
	gologger.Info().Msgf("Rate limiter updated: delay=%v, burst=%d", minDelay, burstSize)
}

// GetRateLimitStatus 获取限流器状态
func (e *Engine) GetRateLimitStatus() (limit rate.Limit, burst int) {
	return e.limiter.Limit(), e.limiter.Burst()
}

// GetConcurrencyStatus 获取并发状态
func (e *Engine) GetConcurrencyStatus() (globalUsed, globalTotal, targetUsed, targetTotal int) {
	globalTotal = cap(e.globalSem)
	globalUsed = len(e.globalSem)

	// 统计所有目标的并发使用情况
	var totalTargetUsed, totalTargetCap int
	e.processes.Range(func(key, value interface{}) bool {
		process := value.(*targetProcess)
		totalTargetUsed += len(process.semaphore)
		totalTargetCap += cap(process.semaphore)
		return true
	})

	return globalUsed, globalTotal, totalTargetUsed, totalTargetCap
}
