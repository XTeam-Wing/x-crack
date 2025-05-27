package brute

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// TestEngine_ConcurrencyControl 测试并发控制
func TestEngine_ConcurrencyControl(t *testing.T) {
	config := &Config{
		TargetConcurrent: 2,
		TaskConcurrent:   3,
		MinDelay:         time.Millisecond * 100,
		Timeout:          time.Second * 5,
		MaxRetries:       1,
		OkToStop:         false,
	}

	ctx := context.Background()
	engine, err := NewEngine(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	var concurrentCount int32
	var maxConcurrent int32

	// 设置自定义回调来监控并发数
	config.CustomCallback = func(item *BruteItem) *BruteResult {
		current := atomic.AddInt32(&concurrentCount, 1)
		defer atomic.AddInt32(&concurrentCount, -1)

		// 更新最大并发数
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		// 模拟处理时间
		time.Sleep(time.Millisecond * 50)

		return &BruteResult{
			Item:    item,
			Success: false,
		}
	}

	// 添加目标
	engine.AddTarget("test", "127.0.0.1", 22)

	// 添加多个任务
	for i := 0; i < 10; i++ {
		item := &BruteItem{
			Type:     "test",
			Target:   "127.0.0.1",
			Port:     22,
			Username: "user",
			Password: "pass",
			Timeout:  config.Timeout,
		}
		if err := engine.Feed(item); err != nil {
			t.Fatalf("Failed to feed item: %v", err)
		}
	}

	// 启动处理
	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	// 验证并发控制
	if maxConcurrent > int32(config.TaskConcurrent) {
		t.Errorf("Max concurrent (%d) exceeded task concurrent limit (%d)", maxConcurrent, config.TaskConcurrent)
	}

	t.Logf("Max concurrent tasks: %d (limit: %d)", maxConcurrent, config.TaskConcurrent)
}

// TestEngine_RateLimit 测试限流器
func TestEngine_RateLimit(t *testing.T) {
	config := &Config{
		TargetConcurrent: 1,
		TaskConcurrent:   1,
		MinDelay:         time.Millisecond * 200, // 200ms 延迟
		Timeout:          time.Second * 5,
		MaxRetries:       1,
		OkToStop:         false,
	}

	ctx := context.Background()
	engine, err := NewEngine(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	var timestamps []time.Time

	// 设置自定义回调来记录时间戳
	config.CustomCallback = func(item *BruteItem) *BruteResult {
		timestamps = append(timestamps, time.Now())
		return &BruteResult{
			Item:    item,
			Success: false,
		}
	}

	// 添加目标
	engine.AddTarget("test", "127.0.0.1", 22)

	// 添加多个任务
	taskCount := 5
	for i := 0; i < taskCount; i++ {
		item := &BruteItem{
			Type:     "test",
			Target:   "127.0.0.1",
			Port:     22,
			Username: "user",
			Password: "pass",
			Timeout:  config.Timeout,
		}
		if err := engine.Feed(item); err != nil {
			t.Fatalf("Failed to feed item: %v", err)
		}
	}

	startTime := time.Now()
	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	totalTime := time.Since(startTime)

	// 验证限流效果
	if len(timestamps) != taskCount {
		t.Fatalf("Expected %d timestamps, got %d", taskCount, len(timestamps))
	}

	// 检查时间间隔
	for i := 1; i < len(timestamps); i++ {
		interval := timestamps[i].Sub(timestamps[i-1])
		if interval < config.MinDelay {
			t.Errorf("Interval %v is less than min delay %v", interval, config.MinDelay)
		}
	}

	// 检查总时间（应该大于 (taskCount-1) * MinDelay）
	expectedMinTime := time.Duration(taskCount-1) * config.MinDelay
	if totalTime < expectedMinTime {
		t.Errorf("Total time %v is less than expected minimum %v", totalTime, expectedMinTime)
	}

	t.Logf("Total time: %v, expected minimum: %v", totalTime, expectedMinTime)
}

// TestEngine_UpdateRateLimit 测试动态更新限流器
func TestEngine_UpdateRateLimit(t *testing.T) {
	config := DefaultConfig()
	ctx := context.Background()
	engine, err := NewEngine(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 检查初始设置
	limit, burst := engine.GetRateLimitStatus()
	if limit != rate.Every(config.MinDelay) {
		t.Errorf("Initial limit mismatch: got %v, expected %v", limit, rate.Every(config.MinDelay))
	}
	if burst != config.TargetConcurrent {
		t.Errorf("Initial burst mismatch: got %d, expected %d", burst, config.TargetConcurrent)
	}

	// 更新限流器
	newDelay := time.Millisecond * 500
	newBurst := 20
	engine.UpdateRateLimit(newDelay, newBurst)

	// 检查更新后的设置
	limit, burst = engine.GetRateLimitStatus()
	if limit != rate.Every(newDelay) {
		t.Errorf("Updated limit mismatch: got %v, expected %v", limit, rate.Every(newDelay))
	}
	if burst != newBurst {
		t.Errorf("Updated burst mismatch: got %d, expected %d", burst, newBurst)
	}
}

// TestEngine_ConcurrencyStatus 测试并发状态监控
func TestEngine_ConcurrencyStatus(t *testing.T) {
	config := &Config{
		TargetConcurrent: 5,
		TaskConcurrent:   3,
		MinDelay:         time.Millisecond * 10,
		Timeout:          time.Second * 5,
	}

	ctx := context.Background()
	engine, err := NewEngine(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 添加目标
	engine.AddTarget("test", "127.0.0.1", 22)

	// 检查并发状态
	globalUsed, globalTotal, targetUsed, targetTotal := engine.GetConcurrencyStatus()

	if globalTotal != config.TargetConcurrent {
		t.Errorf("Global total mismatch: got %d, expected %d", globalTotal, config.TargetConcurrent)
	}
	if globalUsed != 0 {
		t.Errorf("Global used should be 0 initially, got %d", globalUsed)
	}

	t.Logf("Concurrency status - Global: %d/%d, Target: %d/%d",
		globalUsed, globalTotal, targetUsed, targetTotal)
}
