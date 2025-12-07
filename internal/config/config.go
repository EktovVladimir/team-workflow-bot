package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"team-workflow-bot/internal/environment"

	"github.com/spf13/viper"
)

type Config struct {
	Slack      Slack      `mapstructure:"slack"`
	GitHub     GitHub     `mapstructure:"github"`
	GitHubHook GitHubHook `mapstructure:"githubHook"`
	Jira       Jira       `mapstructure:"jira"`
}

type Slack struct {
	AppToken string `mapstructure:"appToken"`
	BotToken string `mapstructure:"botToken"`
}

type GitHub struct {
	Token string `mapstructure:"token"`
}

type GitHubHook struct {
	SecretKey string `mapstructure:"secretKey"`
	BaseURL   string `mapstructure:"baseUrl"`
	Port      int    `mapstructure:"port"`
	Route     string `mapstructure:"route"`
}

type Jira struct {
	Email   string `mapstructure:"email"`
	Token   string `mapstructure:"token"`
	BaseUrl string `mapstructure:"baseUrl"`
}

func Load(appName string) *Config {
	configPath := filepath.Join("configs", fmt.Sprintf("%s.%s.json", appName, environment.Env))

	//TODO setDefaults()

	viper.SetConfigFile(configPath)
	viper.SetConfigType("json")
	viper.SetEnvPrefix(appName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			panic(fmt.Errorf("fatal error reading config file: %w", err))
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("fatal error unmarshaling config: %w", err))
	}

	return &cfg
}
