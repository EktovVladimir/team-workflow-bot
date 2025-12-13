package app

import (
	"context"
	"sync"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/config"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/handlers/codereview"
	"team-workflow-bot/internal/handlers/configurator"
	"team-workflow-bot/internal/integrations/githubflow"
	"team-workflow-bot/internal/integrations/jiraflow"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/logger"
	"team-workflow-bot/internal/services"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v79/github"
	"github.com/sirupsen/logrus"
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

func (a *App) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}

	logrus.Info("Starting Team Workflow Bot...")

	mongo, err := db.ConnectMongo(ctx, a.config)
	if err != nil {
		logrus.Fatalf("Error connecting to Mongo: %v", err)
		return
	}
	defer func() {
		_ = mongo.Disconnect(ctx)
	}()

	mongoDb := mongo.Database(a.config.Mongo.DB)

	if err := db.EnsureIndexes(ctx, mongoDb); err != nil {
		logrus.Fatalf("Error ensuring Mongo indexes: %v", err)
		return
	}

	repository := db.NewRepository(mongoDb)

	err = global.InitGlobalStorageData(ctx, repository)
	if err != nil {
		logrus.Fatalf("Error initializing global storage data: %v", err)
		return
	}

	slackClient := slack.New(
		a.config.Slack.BotToken,
		slack.OptionDebug(global.IsDev),
		slack.OptionAppLevelToken(a.config.Slack.AppToken),
		slack.OptionLog(logger.NewStdLogger(logrus.DebugLevel, "slack-api: ")),
	)

	githubClient := github.NewClient(nil).WithAuthToken(a.config.GitHub.Token)

	tp := jira.BasicAuthTransport{
		Username: a.config.Jira.Email,
		Password: a.config.Jira.Token,
	}
	jiraClient, _ := jira.NewClient(tp.Client(), a.config.Jira.BaseUrl)

	dependencies := &bag.DependenciesBag{
		DB: bag.DatabaseBag{
			Mongo:      mongoDb,
			Repository: repository,
		},
		Client: bag.ClientBag{
			Slack:  slackClient,
			GitHub: githubClient,
			Jira:   jiraClient,
		},
		Services: bag.InfrastructureServiceBag{
			Github: githubflow.NewService(githubClient),
			Jira:   jiraflow.NewService(jiraClient),
			Slack:  slackflow.NewService(slackClient),
		},
	}

	retrieveService := services.NewRetriever(dependencies)

	codeReviewHandler := codereview.NewHandler(dependencies, retrieveService)
	configureHandler := configurator.NewSlackBotConfigurationHandler(dependencies)

	slackListener := slackflow.NewListener(
		slackClient,
		a.config,
		slackflow.WithAnyHandler(codeReviewHandler),
		slackflow.WithCommandHandler(configureHandler),
		slackflow.WithViewSubmissionHandler(configureHandler))

	ghHandler := githubflow.NewHandler(
		a.config,
		githubflow.WithAnyHandler(codeReviewHandler))

	ghListener := githubflow.NewListener(ghHandler, a.config)

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := slackListener.Run(ctx)
		if err != nil {
			logrus.Fatal(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		_ = ghListener.Run(ctx)
	}()

	wg.Wait()
}
