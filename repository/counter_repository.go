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
	SyncBotIDCounter(ctx context.Context) error
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

// GetNextBotID 获取下一个 Bot ID（自增）
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
	botID := fmt.Sprintf("60%08d", result.Seq)
	return botID, nil
}

// SyncBotIDCounter 同步 BotID 计数器（根据现有最大 BotID）
func (r *counterRepository) SyncBotIDCounter(ctx context.Context) error {
	// 获取 bot 集合中最大的 botID
	botCollection := GetCollection(BotCollection)
	
	// 按 botID 降序排序，取第一个
	opts := options.FindOne().SetSort(bson.D{{Key: "botID", Value: -1}})
	var result struct {
		BotID string `bson:"botID"`
	}
	
	err := botCollection.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 没有数据，重置为 0
			filter := bson.M{"_id": "botID"}
			update := bson.M{"$set": bson.M{"seq": 0}}
			opts := options.Update().SetUpsert(true)
			_, err := r.collection.UpdateOne(ctx, filter, update, opts)
			return err
		}
		return fmt.Errorf("查询最大 BotID 失败: %w", err)
	}

	// 解析 BotID 中的序列号（格式：6000000001）
	var maxSeq int64
	if len(result.BotID) >= 10 && result.BotID[:2] == "60" {
		fmt.Sscanf(result.BotID[2:], "%d", &maxSeq)
	}

	// 更新 counter
	filter := bson.M{"_id": "botID"}
	update := bson.M{"$set": bson.M{"seq": maxSeq}}
	opts2 := options.Update().SetUpsert(true)
	_, err = r.collection.UpdateOne(ctx, filter, update, opts2)
	if err != nil {
		return fmt.Errorf("更新计数器失败: %w", err)
	}

	return nil
}

