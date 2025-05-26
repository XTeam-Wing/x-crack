package protocols

import (
	"context"
	"fmt"

	"github.com/XTeam-Wing/x-crack/pkg/brute"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
	dataSourceName := fmt.Sprintf("mongodb://%s:%s@%v:%v/?authMechanism=SCRAM-SHA-1", item.Username, item.Password, item.Target, item.Port)
	clientOptions := options.Client().ApplyURI(dataSourceName)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		result.Error = err
		return result
	}
	defer client.Disconnect(ctx)

	// 尝试ping数据库来验证连接
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Banner = "MongoDB connection successful"
	return result
}
