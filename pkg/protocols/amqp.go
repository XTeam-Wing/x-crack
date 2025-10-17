package protocols

import (
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	amqp "github.com/rabbitmq/amqp091-go"
)

func AMQPBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}
	// 创建带超时的 context
	var target string
	if item.Username == "" && item.Password == "" {
		target = fmt.Sprintf("amqp://%s:%d", item.Target, item.Port)
	} else if item.Password != "" && item.Username != "" {
		target = fmt.Sprintf("amqp://%s:%s@%s:%d", item.Username, item.Password, item.Target, item.Port)
	} else {
		return result
	}
	conn, err := amqp.Dial(target)
	if err != nil {
		result.Error = err
		return result
	}
	defer conn.Close()
	result.Success = true
	result.Banner = "AMQP authentication successful"
	return result
}
