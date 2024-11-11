package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Node struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"node"`
	
	Database struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	
	Sync struct {
		BatchSize int  `mapstructure:"batch_size"`
		DebugMode  bool `mapstructure:"debug_mode"`
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
