package openim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"open-im-server/bot-system/common/types"
)

type Client struct {
	BaseURL     string
	Secret      string
	AdminUserID string
	adminToken  string
	HTTPClient  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:     baseURL,
		Secret:      "openIM123", // 默认 secret，可以通过环境变量覆盖
		AdminUserID: "imAdmin",   // 默认管理员 userID
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// 设置 Secret
func (c *Client) SetSecret(secret string) {
	c.Secret = secret
}

// 设置管理员 UserID
func (c *Client) SetAdminUserID(adminUserID string) {
	c.AdminUserID = adminUserID
}

// 获取管理员 Token
func (c *Client) GetAdminToken() (string, error) {
	// 如果已有 token，直接返回
	if c.adminToken != "" {
		return c.adminToken, nil
	}

	// 获取 admin token
	url := fmt.Sprintf("%s/auth/get_admin_token", c.BaseURL)
	payload := map[string]interface{}{
		"secret":     c.Secret,
		"platformID": 5,
		"userID":     c.AdminUserID,
	}

	resp, err := c.doRequest("POST", url, payload, "")
	if err != nil {
		return "", fmt.Errorf("获取管理员Token失败: %w", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	token, ok := data["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	// 缓存 token
	c.adminToken = token
	return token, nil
}

// 初始化管理员 Token（启动时调用）
func (c *Client) InitAdminToken() error {
	token, err := c.GetAdminToken()
	if err != nil {
		return err
	}
	fmt.Printf("管理员 Token 初始化成功: %s...\n", token[:20])
	return nil
}

// 注册用户（使用 admin token）
func (c *Client) RegisterUser(userID, nickname, faceURL string) error {
	url := fmt.Sprintf("%s/user/user_register", c.BaseURL)

	payload := map[string]interface{}{
		"secret": c.Secret,
		"users": []map[string]string{
			{
				"userID":   userID,
				"nickname": nickname,
				"faceURL":  faceURL,
			},
		},
	}

	// 获取 admin token
	adminToken, err := c.GetAdminToken()
	if err != nil {
		return err
	}

	_, err = c.doRequest("POST", url, payload, adminToken)
	return err
}

// 获取用户 Token（使用 admin token）
func (c *Client) GetUserToken(userID string, platformID int) (string, error) {
	url := fmt.Sprintf("%s/auth/get_user_token", c.BaseURL)

	payload := map[string]interface{}{
		"secret":     c.Secret,
		"userID":     userID,
		"platformID": platformID,
	}

	// 获取 admin token
	adminToken, err := c.GetAdminToken()
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest("POST", url, payload, adminToken)
	if err != nil {
		return "", err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	token, ok := data["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

// 邀请用户加入群组
func (c *Client) InviteToGroup(groupID string, userIDs []string, token string) error {
	url := fmt.Sprintf("%s/group/invite_user_to_group", c.BaseURL)

	payload := map[string]interface{}{
		"groupID":        groupID,
		"invitedUserIDs": userIDs,
		"reason":         "添加 Bot",
	}

	_, err := c.doRequest("POST", url, payload, token)
	return err
}

// 踢出群成员
func (c *Client) KickFromGroup(groupID string, userIDs []string, token string) error {
	url := fmt.Sprintf("%s/group/kick_group", c.BaseURL)

	payload := map[string]interface{}{
		"groupID":       groupID,
		"kickedUserIDs": userIDs,
		"reason":        "移除 Bot",
	}

	_, err := c.doRequest("POST", url, payload, token)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		// 忽略用户不存在的错误（用户可能已被手动移除）
		if strings.Contains(errStr, "not found") ||
			strings.Contains(errStr, "1101") || // UserIDNotFoundError
			strings.Contains(errStr, "useridnotfounderror") {
			return nil
		}
		return err
	}
	return nil
}

// 发送群消息
func (c *Client) SendGroupMessage(req *types.SendMessageRequest, token string) error {
	url := fmt.Sprintf("%s/msg/send_msg", c.BaseURL)
	_, err := c.doRequest("POST", url, req, token)
	return err
}

// 发送单聊消息
func (c *Client) SendSingleMessage(req *types.SendMessageRequest, token string) error {
	url := fmt.Sprintf("%s/msg/send_msg", c.BaseURL)
	_, err := c.doRequest("POST", url, req, token)
	return err
}

// 通用请求方法
func (c *Client) doRequest(method, url string, payload interface{}, token string) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("operationID", c.generateOperationID())
	if token != "" {
		req.Header.Set("token", token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	// 检查错误码
	if errCode, ok := result["errCode"].(float64); ok && errCode != 0 {
		errMsg, _ := result["errMsg"].(string)
		errDlt, _ := result["errDlt"].(string)
		if errDlt != "" {
			return nil, fmt.Errorf("API error: code=%v, msg=%s, detail=%s", errCode, errMsg, errDlt)
		}
		return nil, fmt.Errorf("API error: code=%v, msg=%s", errCode, errMsg)
	}

	return result, nil
}

// 生成 operationID
func (c *Client) generateOperationID() string {
	return fmt.Sprintf("bot-system-%d", time.Now().UnixNano())
}
