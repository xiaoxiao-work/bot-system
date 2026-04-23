package service

import (
	"context"
	"fmt"
	"log"

	"open-im-server/bot-system/config"
	"open-im-server/bot-system/model"
	"open-im-server/bot-system/repository"
)

var (
	groupBotRepo repository.GroupBotRepository
)

// InitGroupService 初始化群组服务
func InitGroupService() {
	groupBotRepo = repository.NewGroupBotRepository()
}

// GetGroupBots 获取群组的 Bot 列表
func GetGroupBots(ctx context.Context, groupID string) ([]*model.Bot, error) {
	// 获取群组的 Bot 关联
	groupBots, err := groupBotRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// 获取 Bot 详情
	botList := make([]*model.Bot, 0, len(groupBots))
	for _, gb := range groupBots {
		bot, err := GetBot(ctx, gb.BotID)
		if err != nil {
			log.Printf("获取 Bot %s 失败: %v", gb.BotID, err)
			continue
		}
		if bot != nil {
			botList = append(botList, bot)
		}
	}
	return botList, nil
}

// GetGroupBotIDs 获取群组订阅的 Bot ID 列表
func GetGroupBotIDs(ctx context.Context, groupID string) ([]string, error) {
	groupBots, err := groupBotRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	botIDs := make([]string, 0, len(groupBots))
	for _, gb := range groupBots {
		botIDs = append(botIDs, gb.BotID)
	}
	return botIDs, nil
}

// IsBotSubscribed 检查 Bot 是否已订阅群组
func IsBotSubscribed(ctx context.Context, groupID, botID string) (bool, error) {
	return groupBotRepo.Exists(ctx, groupID, botID)
}

// SubscribeBot 订阅 Bot 到群组
func SubscribeBot(ctx context.Context, groupID, botID, operatorID string) error {
	subscription := &model.GroupBotSubscription{
		GroupID:      groupID,
		BotID:        botID,
		SubscribedBy: operatorID,
		Status:       1, // 启用
	}
	return groupBotRepo.Create(ctx, subscription)
}

// UnsubscribeBot 取消订阅 Bot
func UnsubscribeBot(ctx context.Context, groupID, botID string) error {
	return groupBotRepo.Delete(ctx, groupID, botID)
}

// InviteBotToGroup 邀请 Bot 加入群组
func InviteBotToGroup(groupID, botID, operatorID string) error {
	// 使用 admin token 获取操作者的 token
	operatorToken, err := config.OpenIMClient.GetUserToken(operatorID, 5)
	if err != nil {
		return fmt.Errorf("获取操作者Token失败: %w", err)
	}

	// 使用操作者的 token 邀请 Bot 加入群组
	err = config.OpenIMClient.InviteToGroup(groupID, []string{botID}, operatorToken)
	if err != nil {
		return fmt.Errorf("邀请Bot加入群组失败: %w", err)
	}

	log.Printf("Bot %s 已加入群组 %s (操作者: %s)", botID, groupID, operatorID)
	return nil
}

// KickBotFromGroup 将 Bot 从群组移除
func KickBotFromGroup(groupID, botID, operatorID string) error {
	// 使用 admin token 获取操作者的 token
	operatorToken, err := config.OpenIMClient.GetUserToken(operatorID, 5)
	if err != nil {
		return fmt.Errorf("获取操作者Token失败: %w", err)
	}

	// 使用操作者的 token 踢出 Bot
	err = config.OpenIMClient.KickFromGroup(groupID, []string{botID}, operatorToken)
	if err != nil {
		return fmt.Errorf("踢出Bot失败: %w", err)
	}

	log.Printf("Bot %s 已从群组 %s 移除 (操作者: %s)", botID, groupID, operatorID)
	return nil
}
