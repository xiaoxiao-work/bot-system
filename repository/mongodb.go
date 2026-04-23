package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"open-im-server/bot-system/config"
)

var (
	// MongoDB 客户端
	mongoClient *mongo.Client
	// 数据库实例
	database *mongo.Database
)

// InitMongoDB 初始化 MongoDB 连接
func InitMongoDB() error {
	cfg := config.Global.MongoDB

	// 构建连接字符串
	uri := fmt.Sprintf("%s/%s?authSource=admin", cfg.URI, cfg.Database)

	// 创建客户端选项
	clientOptions := options.Client().
		ApplyURI(uri).
		SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		}).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetConnectTimeout(30 * time.Second).      // 连接超时 30 秒
		SetServerSelectionTimeout(30 * time.Second) // 服务器选择超时 30 秒

	// 连接到 MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("Ping MongoDB 失败: %w", err)
	}

	mongoClient = client
	database = client.Database(cfg.Database)

	return nil
}

// GetDatabase 获取数据库实例
func GetDatabase() *mongo.Database {
	return database
}

// GetCollection 获取集合
func GetCollection(name string) *mongo.Collection {
	return database.Collection(name)
}

// Close 关闭 MongoDB 连接
func Close() error {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return mongoClient.Disconnect(ctx)
	}
	return nil
}
