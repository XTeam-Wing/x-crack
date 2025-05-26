package brute_test

import (
	"context"
	"testing"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

func TestNewEngine(t *testing.T) {
	ctx := context.Background()
	config := brute.DefaultConfig()

	engine, err := brute.NewEngine(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Engine is nil")
	}
}

func TestBuilder(t *testing.T) {
	ctx := context.Background()

	var results []*brute.BruteResult
	resultCallback := func(result *brute.BruteResult) {
		results = append(results, result)
	}

	builder := brute.NewBuilder(ctx).
		WithTarget("ssh", "127.0.0.1", 22).
		WithUserDict([]string{"test"}).
		WithPassDict([]string{"test"}).
		WithResultCallback(resultCallback).
		WithTimeout(time.Second * 5)

	engine, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Engine is nil")
	}
}

func TestQuickBrute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var results []*brute.BruteResult
	resultCallback := func(result *brute.BruteResult) {
		results = append(results, result)
	}

	err := brute.QuickBrute(ctx, "ssh", "127.0.0.1", 22,
		[]string{"test"}, []string{"test"}, resultCallback)

	// 这个测试预期会失败，因为通常没有SSH服务在测试环境
	if err != nil {
		t.Logf("Expected error for non-existent SSH service: %v", err)
	}
}

func TestBatchBrute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	targets := []brute.Target{
		{Type: "ssh", Host: "127.0.0.1", Port: 22},
		{Type: "mysql", Host: "127.0.0.1", Port: 3306},
	}

	var results []*brute.BruteResult
	resultCallback := func(result *brute.BruteResult) {
		results = append(results, result)
	}

	err := brute.BatchBrute(ctx, targets,
		[]string{"test"}, []string{"test"}, resultCallback)

	// 这个测试预期会失败，因为通常没有这些服务在测试环境
	if err != nil {
		t.Logf("Expected error for non-existent services: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	// 测试有效配置
	validConfig := brute.DefaultConfig()
	engine, err := brute.NewEngine(context.Background(), validConfig)
	if err != nil {
		t.Fatalf("Valid config should not fail: %v", err)
	}
	if engine == nil {
		t.Fatal("Engine should not be nil for valid config")
	}

	// 测试无效配置
	invalidConfig := brute.DefaultConfig()
	invalidConfig.TargetConcurrent = 0

	_, err = brute.NewEngine(context.Background(), invalidConfig)
	if err == nil {
		t.Fatal("Invalid config should fail")
	}
}

func TestBruteResult(t *testing.T) {
	item := &brute.BruteItem{
		Type:     "ssh",
		Target:   "127.0.0.1",
		Username: "test",
		Password: "test",
		Port:     22,
	}

	result := &brute.BruteResult{
		Item:    item,
		Success: true,
	}

	str := result.String()
	if str == "" {
		t.Fatal("Result string should not be empty")
	}

	if !contains(str, "SUCCESS") {
		t.Fatal("Result string should contain SUCCESS")
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					indexOf(str, substr) >= 0)))
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
