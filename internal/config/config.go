package config

type Config struct {
	IsDebug    bool
	Slack      Slack
	GitHub     GitHub
	GitHubHook GitHubHook
}

type Slack struct {
	AppToken string
	BotToken string
}

type GitHub struct {
	Token string
}

type GitHubHook struct {
	SecretKey string
	BaseURL   string
	Port      int
	Route     string
}

func LoadConfig() *Config {
	//TODO environment variables or config files
	return &Config{
		IsDebug: true,
	}
}
