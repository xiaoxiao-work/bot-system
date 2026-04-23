package types

import (
	"encoding/json"
)

// OpenIM Webhook 消息结构
type WebhookMessage struct {
	CallbackCommand string      `json:"callbackCommand"`
	SendID          string      `json:"sendID"`
	RecvID          string      `json:"recvID"`
	GroupID         string      `json:"groupID"`
	ContentType     int         `json:"contentType"`
	SessionType     int         `json:"sessionType"` // 1=单聊, 2=群聊
	Content         interface{} `json:"content"`     // 可能是 string 或 map
	ClientMsgID     string      `json:"clientMsgID"`
	ServerMsgID     string      `json:"serverMsgID"`
	SendTime        int64       `json:"sendTime"`
	SenderNickname  string      `json:"senderNickname"`
	SenderFaceURL   string      `json:"senderFaceURL"`
}

// GetContentMap 获取 Content 的 map 形式
func (m *WebhookMessage) GetContentMap() (map[string]interface{}, error) {
	switch v := m.Content.(type) {
	case map[string]interface{}:
		return v, nil
	case string:
		// 如果是字符串，尝试解析为 JSON
		var contentMap map[string]interface{}
		if err := json.Unmarshal([]byte(v), &contentMap); err != nil {
			return nil, err
		}
		return contentMap, nil
	default:
		return nil, nil
	}
}

// GetContentString 获取 Content 的字符串形式
func (m *WebhookMessage) GetContentString() string {
	switch v := m.Content.(type) {
	case string:
		return v
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok {
			return text
		}
		// 尝试序列化为 JSON
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
	}
	return ""
}

// 发送消息请求
type SendMessageRequest struct {
	SendID            string                 `json:"sendID"`
	RecvID            string                 `json:"recvID,omitempty"`
	GroupID           string                 `json:"groupID,omitempty"`
	SenderNickname    string                 `json:"senderNickname,omitempty"`    // 发送者昵称
	SenderFaceURL     string                 `json:"senderFaceURL,omitempty"`     // 发送者头像
	SenderPlatformID  int                    `json:"senderPlatformID"`
	ContentType       int                    `json:"contentType"`
	SessionType       int                    `json:"sessionType"`
	Content           map[string]interface{} `json:"content"`
	OfflinePushInfo   *OfflinePushInfo       `json:"offlinePushInfo,omitempty"`
}

type OfflinePushInfo struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Ex    string `json:"ex"`
}

// OpenIM API 响应
type OpenIMResponse struct {
	ErrCode int         `json:"errCode"`
	ErrMsg  string      `json:"errMsg"`
	Data    interface{} `json:"data,omitempty"`
}
