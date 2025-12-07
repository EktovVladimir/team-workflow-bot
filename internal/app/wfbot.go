package app

import (
	"context"
	"sync"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/codereview"
	"team-workflow-bot/internal/environment"
	"team-workflow-bot/internal/githubflow"
	"team-workflow-bot/internal/slackflow"

	"log"
	"team-workflow-bot/internal/config"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v79/github"
	"github.com/slack-go/slack"
)

type App struct {
	config *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Start(ctx context.Context) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	log.Println("Starting Team Workflow Bot...")

	slackClient := slack.New(
		a.config.Slack.BotToken,
		slack.OptionDebug(environment.IsDev),
		slack.OptionAppLevelToken(a.config.Slack.AppToken),
	)

	githubClient := github.NewClient(nil).WithAuthToken(a.config.GitHub.Token)

	tp := jira.BasicAuthTransport{
		Username: a.config.Jira.Email,
		Password: a.config.Jira.Token,
	}
	jiraClient, _ := jira.NewClient(tp.Client(), a.config.Jira.BaseUrl)

	dependencies := bag.DependenciesBag{
		Client: bag.ClientsBag{
			Slack:  slackClient,
			GitHub: githubClient,
			Jira:   jiraClient,
		},
	}

	codeReviewHandler := codereview.NewHandler(dependencies)

	slackListener := slackflow.NewListener(
		slackClient,
		a.config,
		slackflow.WithAnyHandler(codeReviewHandler))

	ghHandler := githubflow.NewHandler(
		a.config,
		githubflow.WithAnyHandler(codeReviewHandler))

	ghListener := githubflow.NewListener(ghHandler, a.config)

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := slackListener.Run(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		_ = ghListener.Run(ctx)
	}()

	return wg
}
