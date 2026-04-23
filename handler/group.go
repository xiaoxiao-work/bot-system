package handler

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"open-im-server/bot-system/common/response"
	"open-im-server/bot-system/service"
)

// SubscribeBot 订阅 Bot 到群组
func SubscribeBot(c *gin.Context) {
	groupID := c.Param("groupID")
	ctx := c.Request.Context()

	var req struct {
		BotID      string `json:"botID"`
		OperatorID string `json:"operatorID"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFromErr(c, response.ErrBadRequest, "请求参数错误", err)
		return
	}

	// 确保 Bot 用户在 OpenIM 中存在
	err := service.EnsureBotUserExists(ctx, req.BotID)
	if err != nil {
		log.Printf("确保 Bot 用户存在失败: %v", err)
		response.ErrorFromErr(c, response.ErrRegisterFailed, "确保 Bot 用户存在失败", err)
		return
	}

	// 检查是否已订阅
	subscribed, err := service.IsBotSubscribed(ctx, groupID, req.BotID)
	if err != nil {
		response.ErrorFromErr(c, response.ErrInternalServer, "检查订阅状态失败", err)
		return
	}
	if subscribed {
		response.SuccessWithMsg(c, "该 Bot 已在群组中", nil)
		return
	}

	// 邀请 Bot 加入群组
	err = service.InviteBotToGroup(groupID, req.BotID, req.OperatorID)
	if err != nil {
		log.Printf("邀请 Bot 加入群组失败: %v", err)
		response.ErrorFromErr(c, response.ErrInviteFailed, "邀请 Bot 加入群组失败", err)
		return
	}

	// 记录订阅关系
	err = service.SubscribeBot(ctx, groupID, req.BotID, req.OperatorID)
	if err != nil {
		log.Printf("保存订阅关系失败: %v", err)
		response.ErrorFromErr(c, response.ErrInternalServer, "保存订阅关系失败", err)
		return
	}

	response.SuccessWithMsg(c, fmt.Sprintf("成功添加 Bot %s 到群组", req.BotID), gin.H{
		"botID":   req.BotID,
		"groupID": groupID,
	})
}

// UnsubscribeBot 取消订阅 Bot
func UnsubscribeBot(c *gin.Context) {
	groupID := c.Param("groupID")
	ctx := c.Request.Context()

	var req struct {
		BotID      string `json:"botID"`
		OperatorID string `json:"operatorID"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFromErr(c, response.ErrBadRequest, "请求参数错误", err)
		return
	}

	// 检查 Bot 是否存在
	_, err := service.GetBot(ctx, req.BotID)
	if err != nil {
		response.ErrorFromErr(c, response.ErrBotNotFound, "Bot 不存在", err)
		return
	}

	// 踢出 Bot
	err = service.KickBotFromGroup(groupID, req.BotID, req.OperatorID)
	if err != nil {
		response.ErrorFromErr(c, response.ErrKickFailed, "移除 Bot 失败", err)
		return
	}

	// 删除订阅关系
	err = service.UnsubscribeBot(ctx, groupID, req.BotID)
	if err != nil {
		response.ErrorFromErr(c, response.ErrInternalServer, "删除订阅关系失败", err)
		return
	}

	response.SuccessWithMsg(c, fmt.Sprintf("成功移除 Bot %s", req.BotID), gin.H{
		"botID":   req.BotID,
		"groupID": groupID,
	})
}

// GetGroupBots 获取群组的 Bot 列表
func GetGroupBots(c *gin.Context) {
	groupID := c.Param("groupID")
	ctx := c.Request.Context()

	botList, err := service.GetGroupBots(ctx, groupID)
	if err != nil {
		response.ErrorFromErr(c, response.ErrInternalServer, "获取群组 Bot 列表失败", err)
		return
	}
	response.SuccessResponse(c, botList)
}
