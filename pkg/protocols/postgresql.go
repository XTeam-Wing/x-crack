package protocols

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	_ "github.com/lib/pq"
)

// PostgreSQLBrute PostgreSQL爆破
func PostgreSQLBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable connect_timeout=%d",
		item.Target, item.Port, item.Username, item.Password, int(timeout.Seconds()))

	db, err := sql.Open("postgres", psqlInfo)
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
	result.Banner = "PostgreSQL connection successful"
	return result
}
