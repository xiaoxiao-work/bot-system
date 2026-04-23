package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"open-im-server/bot-system/config"
	"open-im-server/bot-system/handler"
	"open-im-server/bot-system/repository"
	"open-im-server/bot-system/service"
)

func main() {
	// 加载配置文件
	if err := config.LoadConfig("config/config.yml"); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	log.Println("配置文件加载成功")

	// 初始化 MongoDB
	log.Println("正在连接 MongoDB...")
	if err := repository.InitMongoDB(); err != nil {
		log.Fatalf("连接 MongoDB 失败: %v", err)
	}
	log.Println("MongoDB 连接成功")

	// 初始化数据库索引
	ctx := context.Background()
	if err := repository.InitIndexes(ctx); err != nil {
		log.Printf("初始化索引失败: %v", err)
	}

	// 初始化 OpenIM 客户端
	log.Println("正在初始化 OpenIM 客户端...")
	if err := config.InitOpenIMClient(); err != nil {
		log.Fatalf("初始化 OpenIM 客户端失败: %v", err)
	}
	log.Println("OpenIM 客户端初始化成功")

	// 初始化服务
	service.InitBotService()
	service.InitGroupService()
	log.Println("服务初始化完成")

	// 加载 Bot 缓存
	if err := service.LoadBotCache(ctx); err != nil {
		log.Printf("加载 Bot 缓存失败: %v", err)
	} else {
		log.Println("Bot 缓存加载完成")
	}

	// 初始化 Gin
	r := gin.Default()

	// Webhook 接收（OpenIM 回调）
	r.POST("/webhook/callbackAfterSendGroupMsgCommand", handler.HandleWebhookAfterSendGroupMsgCommand)

	// Bot 管理 API
	api := r.Group("/api")
	{
		api.GET("/bots", handler.ListBots)
		api.POST("/group/:groupID/subscribe_bot", handler.SubscribeBot)
		api.POST("/group/:groupID/unsubscribe_bot", handler.UnsubscribeBot)
		api.GET("/group/:groupID/bots", handler.GetGroupBots)

		// Bot 发送消息回调接口
		api.POST("/bot/send_message", handler.HandleBotSendMessage)
	}

	// 启动服务
	go func() {
		log.Printf("Bot Manager 启动在 %s", config.Global.Server.Port)
		if err := r.Run(config.Global.Server.Port); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")

	// 关闭 MongoDB 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repository.Close(); err != nil {
		log.Printf("关闭 MongoDB 连接失败: %v", err)
	}

	log.Println("服务已关闭")
}
