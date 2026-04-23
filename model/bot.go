package model

import "time"

// Bot 配置
type Bot struct {
	BotID       string    `json:"botID" bson:"botID"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	FaceURL     string    `json:"faceURL" bson:"faceURL"`
	WebhookURL  string    `json:"webhookURL" bson:"webhookURL"`
	Secret      string    `json:"secret" bson:"secret"`         // Bot 密钥，用于验证回调
	Commands    []string  `json:"commands" bson:"commands"`
	IsPublic    bool      `json:"isPublic" bson:"isPublic"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
}

// 群组订阅关系
type GroupBotSubscription struct {
	GroupID      string    `json:"groupID" bson:"groupID"`
	BotID        string    `json:"botID" bson:"botID"`
	SubscribedBy string    `json:"subscribedBy" bson:"subscribedBy"`
	SubscribedAt time.Time `json:"subscribedAt" bson:"subscribedAt"`
	Status       int       `json:"status" bson:"status"` // 0=禁用, 1=启用
}
