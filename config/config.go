package config

import "open-im-server/bot-system/common/openim"

// Config 全局配置
type Config struct {
	OpenIMAPI    string
	OpenIMSecret string
	ServerPort   string
	ServerURL    string // Bot Manager 的外部访问地址
	MongoDB      MongoDBConfig
}

// MongoDBConfig MongoDB 配置
type MongoDBConfig struct {
	URI      string
	Database string
	Username string
	Password string
}

// Global 全局配置实例
var Global = &Config{
	OpenIMAPI:    "http://192.168.31.37:10002",
	OpenIMSecret: "openIM123",
	ServerPort:   ":10006",
	ServerURL:    "http://192.168.31.129:10006", // Bot Manager 的外部访问地址
	MongoDB: MongoDBConfig{
		URI:      "mongodb://192.168.31.37:27017",
		Database: "openim_v3",
		Username: "root",
		Password: "openIM123",
	},
}

// OpenIMClient 全局 OpenIM 客户端
var OpenIMClient *openim.Client

// InitOpenIMClient 初始化 OpenIM 客户端
func InitOpenIMClient() error {
	OpenIMClient = openim.NewClient(Global.OpenIMAPI)
	OpenIMClient.SetSecret(Global.OpenIMSecret)
	return OpenIMClient.InitAdminToken()
}
