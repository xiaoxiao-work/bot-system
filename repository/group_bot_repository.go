package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"open-im-server/bot-system/model"
)

const (
	GroupBotCollection = "group_bot"
)

// GroupBotRepository 群组 Bot 关联数据访问接口
type GroupBotRepository interface {
	Create(ctx context.Context, groupBot *model.GroupBotSubscription) error
	Delete(ctx context.Context, groupID, botID string) error
	FindByGroupID(ctx context.Context, groupID string) ([]*model.GroupBotSubscription, error)
	FindByBotID(ctx context.Context, botID string) ([]*model.GroupBotSubscription, error)
	Exists(ctx context.Context, groupID, botID string) (bool, error)
}

type groupBotRepository struct {
	collection *mongo.Collection
}

// NewGroupBotRepository 创建群组 Bot 仓库实例
func NewGroupBotRepository() GroupBotRepository {
	return &groupBotRepository{
		collection: GetCollection(GroupBotCollection),
	}
}

// Create 创建群组 Bot 关联
func (r *groupBotRepository) Create(ctx context.Context, groupBot *model.GroupBotSubscription) error {
	groupBot.SubscribedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, groupBot)
	return err
}

// Delete 删除群组 Bot 关联
func (r *groupBotRepository) Delete(ctx context.Context, groupID, botID string) error {
	filter := bson.M{
		"groupID": groupID,
		"botID":   botID,
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindByGroupID 根据群组 ID 查找所有 Bot 关联
func (r *groupBotRepository) FindByGroupID(ctx context.Context, groupID string) ([]*model.GroupBotSubscription, error) {
	filter := bson.M{"groupID": groupID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groupBots []*model.GroupBotSubscription
	if err = cursor.All(ctx, &groupBots); err != nil {
		return nil, err
	}
	return groupBots, nil
}

// FindByBotID 根据 Bot ID 查找所有群组关联
func (r *groupBotRepository) FindByBotID(ctx context.Context, botID string) ([]*model.GroupBotSubscription, error) {
	filter := bson.M{"botID": botID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groupBots []*model.GroupBotSubscription
	if err = cursor.All(ctx, &groupBots); err != nil {
		return nil, err
	}
	return groupBots, nil
}

// Exists 检查群组 Bot 关联是否存在
func (r *groupBotRepository) Exists(ctx context.Context, groupID, botID string) (bool, error) {
	filter := bson.M{
		"groupID": groupID,
		"botID":   botID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
