package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/XTeam-Wing/x-crack/pkg/protocols"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

func main() {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelDebug)
	// 创建配置
	config := &brute.Config{
		TargetConcurrent: 3,                      // 全局最大并发数
		TaskConcurrent:   2,                      // 单个目标的最大并发数
		MinDelay:         time.Millisecond * 500, // 最小延迟500ms
		Timeout:          time.Second * 30,       // 超时时间
		MaxRetries:       1,                      // 最大重试次数
		OkToStop:         false,                  // 成功后不停止
		ShowProgress:     true,                   // 显示进度
	}
	protocols.RegisterAllProtocols()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	targets := []struct {
		service string
		host    string
		port    int
	}{
		{"telnet", "127.0.0.1", 23},
		// {"ftp", "127.0.0.1", 23},
		// {"mysql", "127.0.0.1", 3306},
	}
	// 添加爆破任务
	usernames := []string{"ftp"}
	passwords := []string{"123456", "admin", "password"}
	tt := make([]brute.Target, 0, len(targets))
	for _, target := range targets {
		tt = append(tt, brute.Target{
			Type: target.service,
			Host: target.host,
			Port: target.port,
		})
	}
	resultCallback := func(result *brute.BruteResult) {
		if result.Success {
			fmt.Printf("[SUCCESS] %s://%s:%d - %s:%s\n", result.Item.Type, result.Item.Target, result.Item.Port, result.Item.Username, result.Item.Password)
		}
	}
	err := brute.BatchBruteWithConfig(ctx, tt, usernames, passwords, resultCallback, config)
	if err != nil {
		log.Fatalf("Failed to start brute force: %v", err)
	}
}
