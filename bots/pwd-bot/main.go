package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// WebhookPayload Bot Manager 发送的 Webhook 数据
type WebhookPayload struct {
	Message           Message `json:"message"`
	BotID             string  `json:"botID"`
	BotSecret         string  `json:"botSecret"`         // Bot 密钥，用于回调验证
	ReplyCallbackURL  string  `json:"replyCallbackURL"` // 回复消息的回调地址
}

// Message 消息结构
type Message struct {
	SendID         string      `json:"sendID"`
	GroupID        string      `json:"groupID"`
	ContentType    int         `json:"contentType"`
	SessionType    int         `json:"sessionType"`
	Content        interface{} `json:"content"`
	SenderNickname string      `json:"senderNickname"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	SendID           string                 `json:"sendID"`
	RecvID           string                 `json:"recvID,omitempty"`
	GroupID          string                 `json:"groupID,omitempty"`
	SenderPlatformID int                    `json:"senderPlatformID"`
	ContentType      int                    `json:"contentType"`
	SessionType      int                    `json:"sessionType"`
	Content          map[string]interface{} `json:"content"`
}

func main() {
	r := gin.Default()

	r.POST("/webhook", handleWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("PWD Bot 启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

func handleWebhook(c *gin.Context) {
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("解析请求失败: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Printf("收到消息: 发送者=%s, 群组=%s", payload.Message.SenderNickname, payload.Message.GroupID)

	// 解析消息内容
	content := getContentString(payload.Message.Content)
	log.Printf("消息内容: %s", content)

	// 检查是否是 /pwd 命令
	if !strings.HasPrefix(content, "/pwd") {
		log.Printf("忽略非 /pwd 命令")
		c.JSON(200, gin.H{"status": "ignored"})
		return
	}

	// 执行 pwd 命令
	cmd := exec.Command("pwd")
	output, err := cmd.CombinedOutput()

	var replyText string
	if err != nil {
		replyText = fmt.Sprintf("❌ 执行失败: %v", err)
	} else {
		replyText = fmt.Sprintf("📍 当前工作目录:\n```\n%s```", strings.TrimSpace(string(output)))
	}

	// 发送回复
	if err := sendReply(payload, replyText); err != nil {
		log.Printf("发送回复失败: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("回复已发送")
	c.JSON(200, gin.H{"status": "success"})
}

func getContentString(content interface{}) string {
	switch v := content.(type) {
	case string:
		// 尝试解析 JSON 字符串
		var contentMap map[string]interface{}
		if err := json.Unmarshal([]byte(v), &contentMap); err == nil {
			if text, ok := contentMap["content"].(string); ok {
				return text
			}
		}
		return v
	case map[string]interface{}:
		if text, ok := v["content"].(string); ok {
			return text
		}
	}
	return ""
}

func sendReply(payload WebhookPayload, text string) error {
	req := SendMessageRequest{
		SendID:           payload.BotID,
		GroupID:          payload.Message.GroupID,
		SenderPlatformID: 5,
		ContentType:      101, // 文本消息
		SessionType:      payload.Message.SessionType,
		Content: map[string]interface{}{
			"content": text,
		},
	}

	jsonData, _ := json.Marshal(req)
	log.Printf("发送消息到 Bot Manager: %s", payload.ReplyCallbackURL)
	
	httpReq, err := http.NewRequest("POST", payload.ReplyCallbackURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Bot-Secret", payload.BotSecret) // 携带 Bot Secret 用于验证

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("发送消息失败: HTTP %d", resp.StatusCode)
	}

	return nil
}
