package protocols

import (
	"testing"
	"time"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
)

func TestAMQPBrute(t *testing.T) {
	input := &brute.BruteItem{
		Target:   "127.0.0.1",
		Port:     5672,
		Username: "guest2",
		Password: "guest",
		Timeout:  5 * time.Second,
	}

	result := AMQPBrute(input)

	if !result.Success {
		t.Errorf("Expected success for valid credentials, got failure,%s", result.Error.Error())
		return
	}
	t.Logf("AMQPBrute result: %+v", result)
}

func TestMongoDBBrute(t *testing.T) {
	input := &brute.BruteItem{
		Target:   "127.0.0.1",
		Port:     27018,
		Username: "12",
		Password: "1",
		Timeout:  5 * time.Second,
	}

	result := MongoDBBrute(input)

	if !result.Success {
		t.Errorf("Expected success for valid credentials, got failure,%s", result.Error.Error())
		return
	}
	t.Logf("MongoDBBrute result: %+v", result)
}
