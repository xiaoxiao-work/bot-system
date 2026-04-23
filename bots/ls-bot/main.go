package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// WebhookPayload Bot Manager 发送的 Webhook 数据
type WebhookPayload struct {
	Message          Message `json:"message"`
	BotID            string  `json:"botID"`
	BotSecret        string  `json:"botSecret"`        // Bot 密钥，用于回调验证
	ReplyCallbackURL string  `json:"replyCallbackURL"` // 回复消息的回调地址
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
		port = "8001"
	}

	log.Printf("LS Bot 启动在端口 %s", port)
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

	// 检查是否是 /ls 命令
	if !strings.HasPrefix(content, "/ls") {
		log.Printf("忽略非 /ls 命令")
		c.JSON(200, gin.H{"status": "ignored"})
		return
	}

	// 执行 ls 命令（跨平台支持）
	homeDir, _ := os.UserHomeDir()
	var cmd *exec.Cmd
	var cmdName string

	// 检测操作系统
	if os.PathSeparator == '\\' {
		// Windows
		cmdName = "dir"
		cmd = exec.Command("cmd", "/c", "dir", homeDir)
	} else {
		// Linux/Mac
		cmdName = "ls"
		cmd = exec.Command("ls", "-la", homeDir)
	}

	log.Printf("执行命令: %s %s", cmdName, homeDir)
	output, err := cmd.CombinedOutput()

	var replyText string
	if err != nil {
		replyText = fmt.Sprintf("❌ 执行失败: %v\n输出: %s", err, string(output))
		log.Printf("命令执行失败: %v, 输出: %s", err, string(output))
	} else {
		replyText = fmt.Sprintf("📁 Home 目录内容:\n```\n%s\n```", string(output))
		log.Printf("命令执行成功，输出长度: %d 字节", len(output))
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

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	log.Printf("发送消息到 Bot Manager:")
	log.Printf("  URL: %s", payload.ReplyCallbackURL)
	log.Printf("  Bot ID: %s", payload.BotID)
	log.Printf("  Group ID: %s", payload.Message.GroupID)
	log.Printf("  请求体: %s", string(jsonData))

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

	// 读取响应体
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Bot Manager 响应:")
	log.Printf("  状态码: %d", resp.StatusCode)
	log.Printf("  响应体: %s", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("发送消息失败: HTTP %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应检查错误码
	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err == nil {
		if errCode, ok := apiResp["errCode"].(float64); ok && errCode != 0 {
			errMsg, _ := apiResp["errMsg"].(string)
			errDlt, _ := apiResp["errDlt"].(string)
			return fmt.Errorf("Bot Manager 错误: code=%v, msg=%s, detail=%s", errCode, errMsg, errDlt)
		}
	}

	return nil
}
