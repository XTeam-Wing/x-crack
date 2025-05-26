package protocols

import (
	"strings"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"github.com/go-pg/pg/v10"
)

// PostgreSQLBrute PostgreSQL爆破
func PostgreSQLBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	db := pg.Connect(&pg.Options{
		Addr:     item.Target,
		User:     item.Username,
		Password: item.Password,
		Database: "postgres",
	})
	_, err := db.Exec("select 1")
	if err != nil {
		switch true {
		case strings.Contains(err.Error(), "connect: connection refused"):
			fallthrough
		case strings.Contains(err.Error(), "no pg_hba.conf entry for host"):
			fallthrough
		case strings.Contains(err.Error(), "network unreachable"):
			fallthrough
		case strings.Contains(err.Error(), "i/o timeout"):
			result.Finished = true
			return result
		}
		return result
	}
	result.Success = true
	return result
}
