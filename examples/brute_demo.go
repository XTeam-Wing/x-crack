package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

func main() {
	// 创建配置
	config := &brute.Config{
		TargetConcurrent: 3,                      // 全局最大并发数
		TaskConcurrent:   2,                      // 单个目标的最大并发数
		MinDelay:         time.Millisecond * 500, // 最小延迟500ms
		Timeout:          time.Second * 10,       // 超时时间
		MaxRetries:       1,                      // 最大重试次数
		OkToStop:         false,                  // 成功后不停止
	}

	// 设置自定义回调来演示功能
	config.CustomCallback = func(item *brute.BruteItem) *brute.BruteResult {
		start := time.Now()

		// 模拟爆破过程
		fmt.Printf("[%s] Trying %s://%s:%s@%s:%d\n",
			start.Format("15:04:05.000"),
			item.Type, item.Username, item.Password, item.Target, item.Port)

		// 模拟处理时间
		time.Sleep(time.Millisecond * 100)

		// 模拟成功率（10%的成功率）
		success := (time.Now().UnixNano() % 10) == 0

		return &brute.BruteResult{
			Item:         item,
			Success:      success,
			ResponseTime: time.Since(start),
		}
	}

	ctx := context.Background()
	engine, err := brute.NewEngine(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// 设置结果回调
	engine.SetResultCallback(func(result *brute.BruteResult) {
		status := "FAIL"
		if result.Success {
			status = "SUCCESS"
		}
		fmt.Printf("[RESULT] %s - %s (%v)\n",
			status, result.String(), result.ResponseTime)
	})

	// 添加目标
	targets := []struct {
		service string
		host    string
		port    int
	}{
		{"ssh", "192.168.1.100", 22},
		{"ftp", "192.168.1.101", 21},
		{"mysql", "192.168.1.102", 3306},
	}

	for _, target := range targets {
		engine.AddTarget(target.service, target.host, target.port)
	}

	// 添加爆破任务
	usernames := []string{"admin", "root", "user"}
	passwords := []string{"123456", "admin", "password"}

	for _, target := range targets {
		for _, username := range usernames {
			for _, password := range passwords {
				item := &brute.BruteItem{
					Type:     target.service,
					Target:   target.host,
					Port:     target.port,
					Username: username,
					Password: password,
					Timeout:  config.Timeout,
				}
				if err := engine.Feed(item); err != nil {
					log.Printf("Failed to feed item: %v", err)
				}
			}
		}
	}

	fmt.Printf("=== 开始爆破演示 ===\n")
	fmt.Printf("配置: 全局并发=%d, 目标并发=%d, 延迟=%v\n",
		config.TargetConcurrent, config.TaskConcurrent, config.MinDelay)
	fmt.Printf("目标数量: %d, 任务总数: %d\n",
		engine.GetTargetCount(), len(targets)*len(usernames)*len(passwords))

	// 在另一个goroutine中监控状态
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				globalUsed, globalTotal, targetUsed, targetTotal := engine.GetConcurrencyStatus()
				processedCount := engine.GetProcessedCount()
				fmt.Printf("[STATUS] 处理进度: %d, 并发状态 - 全局: %d/%d, 目标: %d/%d\n",
					processedCount, globalUsed, globalTotal, targetUsed, targetTotal)
			case <-ctx.Done():
				return
			}
		}
	}()

	// 演示动态调整限流器
	go func() {
		time.Sleep(time.Second * 5)
		fmt.Println("[DEMO] 动态调整限流器: 延迟改为300ms, 突发容量改为5")
		engine.UpdateRateLimit(time.Millisecond*300, 5)
	}()

	// 开始爆破
	start := time.Now()
	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	totalTime := time.Since(start)
	processedCount := engine.GetProcessedCount()

	fmt.Printf("\n=== 爆破完成 ===\n")
	fmt.Printf("总用时: %v\n", totalTime)
	fmt.Printf("处理任务数: %d\n", processedCount)
	fmt.Printf("平均每个任务耗时: %v\n", totalTime/time.Duration(processedCount))

	// 获取最终状态
	limit, burst := engine.GetRateLimitStatus()
	fmt.Printf("最终限流器设置: 限制=%v, 突发容量=%d\n", limit, burst)
}
