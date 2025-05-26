package main

import (
	"context"
	"fmt"
	"time"

	"github.com/x/x-crack/pkg/brute"
	"github.com/x/x-crack/pkg/config"
)

func main() {
	// 创建配置
	cfg := &config.Config{
		Timeout:          5 * time.Second,
		TargetConcurrent: 10,
		TaskConcurrent:   20,
	}

	// 创建目标
	targets := []brute.Target{
		{ServiceType: "ssh", Host: "127.0.0.1", Port: 22},
	}

	// 创建 Builder
	builder := brute.NewBuilder(context.Background()).
		WithConfig(cfg).
		WithTargets(targets).
		WithUsernames([]string{"admin"}).
		WithPasswords([]string{"admin123"})

	// 生成 BruteItem 来检查超时值
	items := builder.GenerateBruteItems()
	if len(items) > 0 {
		fmt.Printf("Generated BruteItem:\n")
		fmt.Printf("  Protocol: %s\n", items[0].Protocol)
		fmt.Printf("  Host: %s\n", items[0].Host)
		fmt.Printf("  Port: %d\n", items[0].Port)
		fmt.Printf("  Username: %s\n", items[0].Username)
		fmt.Printf("  Password: %s\n", items[0].Password)
		fmt.Printf("  Timeout: %v\n", items[0].Timeout)
		fmt.Printf("  Config Timeout: %v\n", cfg.Timeout)
	}
}
