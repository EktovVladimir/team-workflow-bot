package app

import (
	"context"
	"sync"
	"team-workflow-bot/internal/githubflow"
	"team-workflow-bot/internal/slackflow"

	"log"
	"team-workflow-bot/internal/config"
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

	slackListener := slackflow.NewListenerAndClient(a.config)

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := slackListener.Run(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
	}()

	ghHandler := githubflow.NewHandler(a.config)
	ghListener := githubflow.NewListener(ghHandler, a.config)

	wg.Add(1)
	go func() {
		defer wg.Done()

		_ = ghListener.Run(ctx)
	}()

	return wg
}
