package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"open-im-server/bot-system/common/response"
	"open-im-server/bot-system/common/types"
	"open-im-server/bot-system/config"
	"open-im-server/bot-system/model"
	"open-im-server/bot-system/service"
)

// HandleWebhookAfterSendGroupMsgCommand 处理 OpenIM Webhook
func HandleWebhookAfterSendGroupMsgCommand(c *gin.Context) {
	// 记录原始请求体用于调试
	bodyBytes, _ := c.GetRawData()

	// 重新设置请求体以便后续解析
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var msg types.WebhookMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		response.ErrorFromErr(c, response.ErrBadRequest, "请求参数错误", err)
		return
	}

	// 获取回调命令（从路径参数或消息体）
	// 使用通配符时，参数会包含前导斜杠，需要去掉
	callbackCommand := c.Param("callbackCommand")
	if callbackCommand != "" && callbackCommand[0] == '/' {
		callbackCommand = callbackCommand[1:]
	}
	if callbackCommand == "" {
		callbackCommand = msg.CallbackCommand
	}

	// 只处理群消息（普通群=2, 超级群=3）
	if msg.SessionType != 2 && msg.SessionType != 3 {
		log.Printf("忽略非群消息，会话类型: %d", msg.SessionType)
		response.SuccessResponse(c, nil)
		return
	}

	// 检查是否是群消息相关的回调
	if callbackCommand != "" &&
		callbackCommand != "callbackAfterSendGroupMsgCommand" &&
		callbackCommand != "callbackBeforeSendGroupMsgCommand" {
		log.Printf("忽略非群消息回调: %s", callbackCommand)
		response.SuccessResponse(c, nil)
		return
	}

	ctx := c.Request.Context()

	// 获取该群订阅的 Bot
	subscribedBots, err := service.GetGroupBotIDs(ctx, msg.GroupID)
	if err != nil {
		log.Printf("获取群组 Bot 列表失败: %v", err)
		response.SuccessResponse(c, nil)
		return
	}

	if len(subscribedBots) == 0 {
		log.Printf("群组 %s 没有订阅任何 Bot", msg.GroupID)
		response.SuccessResponse(c, nil)
		return
	}

	// 转发给所有订阅的 Bot
	for _, botID := range subscribedBots {
		// 忽略 Bot 自己的消息
		if msg.SendID == botID {
			log.Printf("忽略 Bot %s 自己的消息", botID)
			continue
		}

		bot, err := service.GetBot(ctx, botID)
		if err != nil || bot == nil {
			log.Printf("获取 Bot %s 信息失败: %v", botID, err)
			continue
		}

		if bot.WebhookURL == "" {
			log.Printf("Bot %s 没有配置 WebhookURL", botID)
			continue
		}

		// 异步转发
		log.Printf("转发消息到 Bot %s (%s)", bot.Name, bot.WebhookURL)
		go forwardToBot(bot, &msg)
	}

	response.SuccessResponse(c, nil)
}

// forwardToBot 转发消息到 Bot 服务
func forwardToBot(bot *model.Bot, msg *types.WebhookMessage) {
	client := &http.Client{Timeout: 5 * time.Second}

	// 构造转发给 Bot 的消息（包含 secret 用于回调验证）
	payload := map[string]interface{}{
		"message":          msg,
		"botID":            bot.BotID,
		"botSecret":        bot.Secret,                                        // Bot 密钥，用于回调验证
		"replyCallbackURL": config.Global.ServerURL + "/api/bot/send_message", // Bot 回复的回调地址
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := client.Post(bot.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("转发到 %s 失败: %v", bot.Name, err)
		return
	}
	defer resp.Body.Close()
}

// HandleBotSendMessage Bot 发送消息的回调接口
func HandleBotSendMessage(c *gin.Context) {
	var req types.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFromErr(c, response.ErrBadRequest, "请求参数错误", err)
		return
	}

	ctx := c.Request.Context()

	// 获取 Bot 信息（使用缓存，高频调用）
	botInfo, err := service.GetBotInfo(ctx, req.SendID)
	if err != nil {
		log.Printf("获取 Bot 信息失败: %v", err)
		response.ErrorFromErr(c, response.ErrInternalServer, "获取 Bot 信息失败", err)
		return
	}
	if botInfo == nil {
		response.ErrorResponse(c, response.ErrBotNotFound, "Bot 不存在")
		return
	}

	// 验证 Bot Secret（防止伪造）
	botSecret := c.GetHeader("X-Bot-Secret")
	if botSecret == "" {
		// 兼容：也可以从 query 参数获取
		botSecret = c.Query("botSecret")
	}

	if botSecret != botInfo.Secret {
		log.Printf("Bot %s Secret 验证失败: 期望=%s..., 实际=%s", req.SendID, botInfo.Secret[:10], botSecret)
		response.ErrorResponse(c, response.ErrBadRequest, "Bot Secret 验证失败")
		return
	}

	// 填充发送者信息
	req.SenderNickname = botInfo.Name
	req.SenderFaceURL = botInfo.FaceURL

	// 使用管理员 Token 发送消息
	adminToken, err := config.OpenIMClient.GetAdminToken()
	if err != nil {
		log.Printf("获取管理员 Token 失败: %v", err)
		response.ErrorFromErr(c, response.ErrGetTokenFailed, "获取管理员 Token 失败", err)
		return
	}

	// 发送消息到 OpenIM
	err = config.OpenIMClient.SendGroupMessage(&req, adminToken)
	if err != nil {
		log.Printf("发送消息到 OpenIM 失败: %v", err)
		response.ErrorFromErr(c, response.ErrSendMessageFailed, "发送消息失败", err)
		return
	}

	response.SuccessWithMsg(c, "消息发送成功", nil)
}
