package repository

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitIndexes 初始化数据库索引
func InitIndexes(ctx context.Context) error {
	// Bot 集合索引
	botCollection := GetCollection(BotCollection)
	botIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "botID", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "isPublic", Value: 1}},
		},
	}

	_, err := botCollection.Indexes().CreateMany(ctx, botIndexes)
	if err != nil {
		log.Printf("创建 Bot 索引失败: %v", err)
		return err
	}
	log.Println("Bot 集合索引创建成功")

	// GroupBot 集合索引
	groupBotCollection := GetCollection(GroupBotCollection)
	groupBotIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "groupID", Value: 1},
				{Key: "botID", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "groupID", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "botID", Value: 1}},
		},
	}

	_, err = groupBotCollection.Indexes().CreateMany(ctx, groupBotIndexes)
	if err != nil {
		log.Printf("创建 GroupBot 索引失败: %v", err)
		return err
	}
	log.Println("GroupBot 集合索引创建成功")

	return nil
}
