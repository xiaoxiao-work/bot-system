package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"github.com/gin-gonic/gin"

	"open-im-server/bot-system/common/response"
	"open-im-server/bot-system/model"
	"open-im-server/bot-system/service"
)

// ListBots 获取 Bot 列表（系统级 + 指定群组）
func ListBots(c *gin.Context) {
	groupID := c.Query("groupID") // 从查询参数获取 groupID

	if groupID == "" {
		response.ErrorResponse(c, response.ErrBadRequest, "缺少 groupID 参数")
		return
	}

	botList, err := service.GetBotsByGroupID(c.Request.Context(), groupID)
	if err != nil {
		response.ErrorFromErr(c, response.ErrInternalServer, "获取 Bot 列表失败", err)
		return
	}
	response.SuccessResponse(c, botList)
}

// CreateBot 创建 Bot
func CreateBot(c *gin.Context) {
	var req struct {
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description"`
		FaceURL     string             `json:"faceURL"`
		WebhookURL  string             `json:"webhookURL" binding:"required"`
		Commands    []model.BotCommand `json:"commands"`
		CreatorID   string             `json:"creatorID" binding:"required"` // 创建者 ID（必填）
		GroupID     string             `json:"groupID" binding:"required"`   // 所属群组 ID（必填）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFromErr(c, response.ErrBadRequest, "请求参数错误", err)
		return
	}

	ctx := c.Request.Context()

	// 生成随机 Secret
	secret := generateSecret()

	bot := &model.Bot{
		// BotID 由 Service 层自动生成
		Name:        req.Name,
		Description: req.Description,
		FaceURL:     req.FaceURL,
		WebhookURL:  req.WebhookURL,
		Secret:      secret,
		Commands:    req.Commands,
		CreatorID:   req.CreatorID,
		GroupID:     req.GroupID,
	}

	if err := service.CreateBot(ctx, bot); err != nil {
		log.Printf("创建 Bot 失败: %v", err)
		response.ErrorFromErr(c, response.ErrInternalServer, "创建 Bot 失败", err)
		return
	}

	response.SuccessWithMsg(c, "Bot 创建成功", gin.H{
		"botID":  bot.BotID, // 返回自动生成的 BotID
		"secret": secret,    // 返回 secret，用户需要保存
	})
}

// DeleteBot 删除 Bot
func DeleteBot(c *gin.Context) {
	botID := c.Param("botID")

	var req struct {
		CreatorID string `json:"creatorID" binding:"required"` // 创建者 ID（必填）
		GroupID   string `json:"groupID" binding:"required"`   // 群组 ID（必填）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFromErr(c, response.ErrBadRequest, "请求参数错误", err)
		return
	}

	ctx := c.Request.Context()

	if err := service.DeleteBot(ctx, botID, req.CreatorID, req.GroupID); err != nil {
		log.Printf("删除 Bot 失败: %v", err)
		response.ErrorFromErr(c, response.ErrInternalServer, err.Error(), err)
		return
	}

	response.SuccessWithMsg(c, "Bot 删除成功", gin.H{
		"botID": botID,
	})
}

// generateSecret 生成随机密钥
func generateSecret() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
