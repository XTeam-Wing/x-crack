package protocols

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	_ "github.com/go-sql-driver/mysql"
)

// MySQLBrute MySQL爆破
func MySQLBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	if item.Username == "" {
		return result
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), item.Timeout)
	defer cancel()

	// 构建DSN连接字符串，添加更多参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%s&readTimeout=%s&writeTimeout=%s",
		item.Username, item.Password, item.Target, item.Port,
		item.Timeout.String(), item.Timeout.String(), item.Timeout.String())

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		result.Error = fmt.Errorf("failed to create MySQL connection: %w", err)
		return result
	}

	// 确保连接关闭
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	// 设置连接池参数
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0) // 不保持空闲连接
	db.SetConnMaxLifetime(item.Timeout)

	// 使用带上下文的Ping验证连接
	if err := db.PingContext(ctx); err != nil {
		result.Error = fmt.Errorf("failed to connect to MySQL: %w", err)
		return result
	}

	// 执行一个简单的查询来进一步验证
	var version string
	err = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		result.Error = fmt.Errorf("failed to query MySQL: %w", err)
		return result
	}

	result.Success = true
	result.Banner = fmt.Sprintf("MySQL connection successful - %s", version)
	return result
}
