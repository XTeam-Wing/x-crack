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

	timeout := item.Timeout
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%s",
		item.Username, item.Password, item.Target, item.Port, timeout.String())

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		result.Error = err
		return result
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Banner = "MySQL connection successful"
	return result
}
