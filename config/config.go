package config

import (
	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

type Config struct {
	Node struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"node"`

	Database struct {
		DSN      string `mapstructure:"dsn"`
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"database"`

	Sync struct {
		BatchSize    int  `mapstructure:"batch_size"`
		DebugMode    bool `mapstructure:"debug_mode"`
		IsFromConfig bool `mapstructure:"is_from_config"`
		StartHeight  int  `mapstructure:"start_height"`
	} `mapstructure:"sync"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetLogLevel 将字符串日志级别转换为 logger.LogLevel
func (c *Config) GetLogLevel() logger.LogLevel {
	switch c.Database.LogLevel {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info // 默认使用 Info 级别
	}
}
