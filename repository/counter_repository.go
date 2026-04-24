package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CounterCollection = "counter"
)

// CounterRepository 计数器数据访问接口
type CounterRepository interface {
	GetNextBotID(ctx context.Context) (string, error)
}

type counterRepository struct {
	collection *mongo.Collection
}

// NewCounterRepository 创建计数器仓库实例
func NewCounterRepository() CounterRepository {
	return &counterRepository{
		collection: GetCollection(CounterCollection),
	}
}

// GetNextBotID 获取下一个 Bot ID（数据库原子自增）
func (r *counterRepository) GetNextBotID(ctx context.Context) (string, error) {
	filter := bson.M{"_id": "botID"}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		ID  string `bson:"_id"`
		Seq int64  `bson:"seq"`
	}

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("获取自增 BotID 失败: %w", err)
	}

	// 生成 BotID，格式：6000000001, 6000000002, ...
	// 序列号永远递增，即使删除也不会重复使用
	botID := fmt.Sprintf("60%08d", result.Seq)
	return botID, nil
}

