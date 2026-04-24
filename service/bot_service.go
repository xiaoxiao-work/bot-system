package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"open-im-server/bot-system/config"
	"open-im-server/bot-system/model"
	"open-im-server/bot-system/repository"
)

var (
	botRepo      repository.BotRepository
	counterRepo  repository.CounterRepository
	botInfoCache *BotInfoCache // Bot 信息缓存（用于高频场景）
)

// BotInfo 缓存的 Bot 基本信息
type BotInfo struct {
	BotID   string
	Name    string
	FaceURL string
	Secret  string // Bot 密钥，用于验证
}

// BotInfoCache Bot 信息缓存（轻量级，只缓存基本信息）
type BotInfoCache struct {
	mu       sync.RWMutex
	botInfos map[string]*BotInfo // botID -> BotInfo
}

// NewBotInfoCache 创建 Bot 信息缓存
func NewBotInfoCache() *BotInfoCache {
	return &BotInfoCache{
		botInfos: make(map[string]*BotInfo),
	}
}

// Get 获取 Bot 信息（从缓存）
func (c *BotInfoCache) Get(botID string) (*BotInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, ok := c.botInfos[botID]
	return info, ok
}

// Set 设置 Bot 信息
func (c *BotInfoCache) Set(botID string, info *BotInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.botInfos[botID] = info
}

// Delete 删除缓存
func (c *BotInfoCache) Delete(botID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.botInfos, botID)
}

// LoadAll 加载所有 Bot 信息
func (c *BotInfoCache) LoadAll(ctx context.Context, repo repository.BotRepository) error {
	bots, err := repo.FindAll(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空并重新填充
	for k := range c.botInfos {
		delete(c.botInfos, k)
	}

	for _, bot := range bots {
		c.botInfos[bot.BotID] = &BotInfo{
			BotID:   bot.BotID,
			Name:    bot.Name,
			FaceURL: bot.FaceURL,
			Secret:  bot.Secret,
		}
	}

	log.Printf("Bot 信息缓存已加载，共 %d 个 Bot", len(c.botInfos))
	return nil
}

// InitBotService 初始化 Bot 服务
func InitBotService() {
	botRepo = repository.NewBotRepository()
	counterRepo = repository.NewCounterRepository()
	botInfoCache = NewBotInfoCache()
}

// LoadBotCache 加载 Bot 缓存
func LoadBotCache(ctx context.Context) error {
	return botInfoCache.LoadAll(ctx, botRepo)
}

// GetAllBots 获取所有 Bot（直接查数据库，保证数据准确）
func GetAllBots(ctx context.Context) ([]*model.Bot, error) {
	return botRepo.FindAll(ctx)
}

// GetBotsByGroupID 获取系统级 Bot 和指定群组的 Bot
func GetBotsByGroupID(ctx context.Context, groupID string) ([]*model.Bot, error) {
	return botRepo.FindByGroupID(ctx, groupID)
}

// GetBot 获取指定 Bot（直接查数据库，保证数据准确）
func GetBot(ctx context.Context, botID string) (*model.Bot, error) {
	return botRepo.FindByID(ctx, botID)
}

// GetBotInfo 获取 Bot 基本信息（高频调用，使用缓存）
func GetBotInfo(ctx context.Context, botID string) (*BotInfo, error) {
	// 先从缓存查询
	if info, ok := botInfoCache.Get(botID); ok {
		return info, nil
	}

	// 缓存未命中，查询数据库
	bot, err := botRepo.FindByID(ctx, botID)
	if err != nil {
		return nil, err
	}

	if bot == nil {
		return nil, nil
	}

	// 构造 BotInfo 并写入缓存
	info := &BotInfo{
		BotID:   bot.BotID,
		Name:    bot.Name,
		FaceURL: bot.FaceURL,
		Secret:  bot.Secret,
	}
	botInfoCache.Set(botID, info)

	log.Printf("Bot %s 信息已缓存", botID)
	return info, nil
}

// EnsureBotUserExists 确保 Bot 用户在 OpenIM 中存在
func EnsureBotUserExists(ctx context.Context, botID string) error {
	// 直接查数据库获取 Bot 信息
	bot, err := botRepo.FindByID(ctx, botID)
	if err != nil {
		return fmt.Errorf("查询 Bot 失败: %w", err)
	}

	if bot == nil {
		return fmt.Errorf("Bot %s 不存在，请先在数据库中添加", botID)
	}

	// 尝试注册用户到 OpenIM
	err = config.OpenIMClient.RegisterUser(botID, bot.Name, bot.FaceURL)
	if err != nil {
		errStr := strings.ToLower(err.Error())

		// 检查是否是"已存在"错误（正常情况）
		if strings.Contains(errStr, "already") ||
			strings.Contains(errStr, "exist") ||
			strings.Contains(errStr, "registered") ||
			strings.Contains(errStr, "1102") { // OpenIM 的已注册错误码
			return nil
		}

		log.Printf("注册 Bot 用户失败: botID=%s, error=%v", botID, err)
		return fmt.Errorf("注册 Bot 用户失败: %w", err)
	}

	log.Printf("Bot 用户注册成功: %s (%s)", bot.Name, botID)
	return nil
}

// CreateBot 创建 Bot
func CreateBot(ctx context.Context, bot *model.Bot) error {
	// 自动生成 BotID
	botID, err := counterRepo.GetNextBotID(ctx)
	if err != nil {
		return fmt.Errorf("生成 BotID 失败: %w", err)
	}
	bot.BotID = botID

	// 创建 Bot
	if err := botRepo.Create(ctx, bot); err != nil {
		return fmt.Errorf("创建 Bot 失败: %w", err)
	}

	// 注册到 OpenIM
	if err := config.OpenIMClient.RegisterUser(bot.BotID, bot.Name, bot.FaceURL); err != nil {
		errStr := strings.ToLower(err.Error())
		// 忽略"已存在"错误
		if !strings.Contains(errStr, "already") &&
			!strings.Contains(errStr, "exist") &&
			!strings.Contains(errStr, "registered") &&
			!strings.Contains(errStr, "1102") {
			log.Printf("注册 Bot 用户到 OpenIM 失败: %v", err)
			// 不回滚，允许后续重试
		}
	}

	// 更新缓存
	botInfoCache.Set(bot.BotID, &BotInfo{
		BotID:   bot.BotID,
		Name:    bot.Name,
		FaceURL: bot.FaceURL,
		Secret:  bot.Secret,
	})

	log.Printf("Bot 创建成功: %s (%s)", bot.Name, bot.BotID)
	return nil
}

// DeleteBot 删除 Bot（带权限检查）
func DeleteBot(ctx context.Context, botID, creatorID, groupID string) error {
	// 查询 Bot
	bot, err := botRepo.FindByID(ctx, botID)
	if err != nil {
		return fmt.Errorf("查询 Bot 失败: %w", err)
	}
	if bot == nil {
		return fmt.Errorf("Bot %s 不存在", botID)
	}

	// 权限检查：验证创建者和群组
	if bot.CreatorID != creatorID {
		return fmt.Errorf("无权删除此 Bot：创建者不匹配")
	}
	if bot.GroupID != groupID {
		return fmt.Errorf("无权删除此 Bot：群组不匹配")
	}

	// 检查是否有群组订阅
	groupBotRepo := repository.NewGroupBotRepository()
	subscriptions, err := groupBotRepo.FindByBotID(ctx, botID)
	if err != nil {
		return fmt.Errorf("检查订阅关系失败: %w", err)
	}
	if len(subscriptions) > 0 {
		return fmt.Errorf("Bot 正在被 %d 个群组使用，无法删除", len(subscriptions))
	}

	// 删除 Bot
	if err := botRepo.Delete(ctx, botID); err != nil {
		return fmt.Errorf("删除 Bot 失败: %w", err)
	}

	// 删除缓存
	botInfoCache.Delete(botID)

	log.Printf("Bot 删除成功: %s (%s)", bot.Name, botID)
	return nil
}
