package model

import "time"

// BotCommand Bot 命令定义
type BotCommand struct {
	Command     string `json:"command" bson:"command"`         // 命令名称，如 "/ls"
	Description string `json:"description" bson:"description"` // 命令描述，如 "列出目录内容"
}

// Bot 配置
type Bot struct {
	BotID       string       `json:"botID" bson:"botID"`
	Name        string       `json:"name" bson:"name"`
	Description string       `json:"description" bson:"description"`
	FaceURL     string       `json:"faceURL" bson:"faceURL"`
	WebhookURL  string       `json:"webhookURL" bson:"webhookURL"`
	Secret      string       `json:"secret" bson:"secret"`     // Bot 密钥，用于验证回调
	Commands    []BotCommand `json:"commands" bson:"commands"` // Bot 支持的命令列表
	IsPublic    bool         `json:"isPublic" bson:"isPublic"`
	CreatedAt   time.Time    `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt" bson:"updatedAt"`
}

// 群组订阅关系
type GroupBotSubscription struct {
	GroupID      string    `json:"groupID" bson:"groupID"`
	BotID        string    `json:"botID" bson:"botID"`
	SubscribedBy string    `json:"subscribedBy" bson:"subscribedBy"`
	SubscribedAt time.Time `json:"subscribedAt" bson:"subscribedAt"`
	Status       int       `json:"status" bson:"status"` // 0=禁用, 1=启用
}
