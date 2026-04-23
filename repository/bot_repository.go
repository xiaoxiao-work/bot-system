package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"open-im-server/bot-system/model"
)

const (
	BotCollection = "bot"
)

// BotRepository Bot 数据访问接口
type BotRepository interface {
	Create(ctx context.Context, bot *model.Bot) error
	Update(ctx context.Context, bot *model.Bot) error
	FindByID(ctx context.Context, botID string) (*model.Bot, error)
	FindAll(ctx context.Context) ([]*model.Bot, error)
	Exists(ctx context.Context, botID string) (bool, error)
}

type botRepository struct {
	collection *mongo.Collection
}

// NewBotRepository 创建 Bot 仓库实例
func NewBotRepository() BotRepository {
	return &botRepository{
		collection: GetCollection(BotCollection),
	}
}

// Create 创建 Bot
func (r *botRepository) Create(ctx context.Context, bot *model.Bot) error {
	bot.CreatedAt = time.Now()
	bot.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, bot)
	return err
}

// Update 更新 Bot
func (r *botRepository) Update(ctx context.Context, bot *model.Bot) error {
	bot.UpdatedAt = time.Now()
	filter := bson.M{"botID": bot.BotID}
	update := bson.M{"$set": bot}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindByID 根据 ID 查找 Bot
func (r *botRepository) FindByID(ctx context.Context, botID string) (*model.Bot, error) {
	var bot model.Bot
	filter := bson.M{"botID": botID}
	err := r.collection.FindOne(ctx, filter).Decode(&bot)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &bot, err
}

// FindAll 查找所有 Bot
func (r *botRepository) FindAll(ctx context.Context) ([]*model.Bot, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bots []*model.Bot
	if err = cursor.All(ctx, &bots); err != nil {
		return nil, err
	}
	return bots, nil
}

// Delete 删除 Bot
func (r *botRepository) Delete(ctx context.Context, botID string) error {
	filter := bson.M{"botID": botID}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// Exists 检查 Bot 是否存在
func (r *botRepository) Exists(ctx context.Context, botID string) (bool, error) {
	filter := bson.M{"botID": botID}
	count, err := r.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
