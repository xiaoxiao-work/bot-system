package handler

import (
	"github.com/gin-gonic/gin"

	"open-im-server/bot-system/common/response"
	"open-im-server/bot-system/service"
)

// ListBots 获取所有 Bot 列表
func ListBots(c *gin.Context) {
	botList, err := service.GetAllBots(c.Request.Context())
	if err != nil {
		response.ErrorFromErr(c, response.ErrInternalServer, "获取 Bot 列表失败", err)
		return
	}
	response.SuccessResponse(c, botList)
}
