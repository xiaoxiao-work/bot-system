package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"open-im-server/bot-system/common/openim"
)

// Config 全局配置
type Config struct {
	OpenIM  OpenIMConfig  `yaml:"openim"`
	Server  ServerConfig  `yaml:"server"`
	MongoDB MongoDBConfig `yaml:"mongodb"`
}

// OpenIMConfig OpenIM 配置
type OpenIMConfig struct {
	API    string `yaml:"api"`
	Secret string `yaml:"secret"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port             string `yaml:"port"`
	URL              string `yaml:"url"`              // Bot Manager 的外部访问地址
	DefaultBotAvatar string `yaml:"defaultBotAvatar"` // 默认 Bot 头像
}

// MongoDBConfig MongoDB 配置
type MongoDBConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Global 全局配置实例
var Global *Config

// OpenIMClient 全局 OpenIM 客户端
var OpenIMClient *openim.Client

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	Global = &Config{}
	if err := yaml.Unmarshal(data, Global); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// InitOpenIMClient 初始化 OpenIM 客户端
func InitOpenIMClient() error {
	OpenIMClient = openim.NewClient(Global.OpenIM.API)
	OpenIMClient.SetSecret(Global.OpenIM.Secret)
	return OpenIMClient.InitAdminToken()
}
