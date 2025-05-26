package protocols

import (
	"context"
	"fmt"
	"net/url"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBBrute MongoDB爆破
func MongoDBBrute(item *brute.BruteItem) *brute.BruteResult {
	result := &brute.BruteResult{
		Item:    item,
		Success: false,
	}

	timeout := item.Timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// MongoDB连接URI
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/admin?serverSelectionTimeoutMS=%d",
		url.QueryEscape(item.Username), url.QueryEscape(item.Password),
		item.Target, item.Port, int(timeout.Milliseconds()))

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		result.Error = err
		return result
	}
	defer client.Disconnect(ctx)

	// 尝试ping数据库来验证连接
	err = client.Ping(ctx, nil)
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Banner = "MongoDB connection successful"
	return result
}
