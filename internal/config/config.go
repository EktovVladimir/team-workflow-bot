package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"team-workflow-bot/internal/environment"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Slack      Slack      `mapstructure:"slack"`
	GitHub     GitHub     `mapstructure:"github"`
	GitHubHook GitHubHook `mapstructure:"githubHook"`
	Jira       Jira       `mapstructure:"jira"`
	Mongo      Mongo      `mapstructure:"mongo"`
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

type Mongo struct {
	Connection string `mapstructure:"connection"`
	DB         string `mapstructure:"db"`
}

func Load(appName string) *Config {
	configPath := filepath.Join("configs", fmt.Sprintf("%s.%s.json", appName, environment.Env))

	setDefaults()

	_ = godotenv.Load(".env")

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

func setDefaults() {
	//Note: установка дефолтных значений обязательна, иначе не будут подтягиваться env переменные
	viper.SetDefault("slack.appToken", "")
	viper.SetDefault("slack.botToken", "")

	viper.SetDefault("github.token", "")

	viper.SetDefault("githubHook.secretKey", "")
	viper.SetDefault("githubHook.baseUrl", "localhost")
	viper.SetDefault("githubHook.port", 8081)
	viper.SetDefault("githubHook.route", "/gh/hook")

	viper.SetDefault("jira.email", "")
	viper.SetDefault("jira.token", "")
	viper.SetDefault("jira.baseUrl", "https://aviasales.atlassian.net")

	viper.SetDefault("mongo.connection", "mongodb://localhost:27017")
	viper.SetDefault("mongo.db", "wf")
}
