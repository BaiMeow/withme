package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
	Database DatabaseConfig `yaml:"database"`
	Gemini   GeminiConfig   `yaml:"gemini"`
	Tencent  TencentConfig  `yaml:"tencent"`
	X        XConfig        `yaml:"x"`
}

// XConfig X (Twitter) API v2 配置；bearer_token 从 developer.x.com 获取，留空则关闭 X 工具
type XConfig struct {
	BearerToken string `yaml:"bearer_token"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

// LogConfig level: debug / info / warn / error，debug 会打印 gemini 思考过程
type LogConfig struct {
	Level string `yaml:"level"`
}

// DatabaseConfig driver: sqlite（本地）/ mysql（线上）
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type GeminiConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// TencentConfig 腾讯云内容安全（TMS）；secret_id/secret_key 为空则关闭审核
type TencentConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
	BizType   string `yaml:"biz_type"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "./data/withme.db"
	}
	if cfg.Gemini.Model == "" {
		cfg.Gemini.Model = "gemini-2.5-flash"
	}
	if cfg.Tencent.Region == "" {
		cfg.Tencent.Region = "ap-guangzhou"
	}
	if cfg.Tencent.BizType == "" {
		cfg.Tencent.BizType = "default"
	}
	return cfg, nil
}
