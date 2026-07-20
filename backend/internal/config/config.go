package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Frontend FrontendConfig `yaml:"frontend"`
	Gemini   GeminiConfig   `yaml:"gemini"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig driver: sqlite（本地）/ mysql（线上）
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// FrontendConfig 前端构建产物目录，由后端托管
type FrontendConfig struct {
	Dist string `yaml:"dist"`
}

type GeminiConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
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
	applyEnv(cfg)
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "./data/withme.db"
	}
	if cfg.Frontend.Dist == "" {
		cfg.Frontend.Dist = "../frontend/dist"
	}
	if cfg.Gemini.Model == "" {
		cfg.Gemini.Model = "gemini-2.5-flash"
	}
	return cfg, nil
}

// applyEnv 环境变量覆盖 yaml 配置，密钥走环境变量，不入库
func applyEnv(cfg *Config) {
	overrides := []struct {
		env string
		dst *string
	}{
		{"SERVER_PORT", &cfg.Server.Port},
		{"DATABASE_DRIVER", &cfg.Database.Driver},
		{"DATABASE_DSN", &cfg.Database.DSN},
		{"FRONTEND_DIST", &cfg.Frontend.Dist},
		{"GEMINI_API_KEY", &cfg.Gemini.APIKey},
		{"GEMINI_MODEL", &cfg.Gemini.Model},
	}
	for _, o := range overrides {
		if v := os.Getenv(o.env); v != "" {
			*o.dst = v
		}
	}
}
