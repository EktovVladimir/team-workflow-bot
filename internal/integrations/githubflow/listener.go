package githubflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"team-workflow-bot/internal/config"

	"github.com/sirupsen/logrus"
)

type Listener struct {
	handler *Handler
	config  *config.Config
}

func NewListener(handler *Handler, cfg *config.Config) *Listener {
	return &Listener{
		handler: handler,
		config:  cfg,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	httpHandler := l.handler.GetHttpHandler()

	http.HandleFunc(l.config.GitHubHook.Route, httpHandler)

	addr := fmt.Sprintf("%s:%v", l.config.GitHubHook.BaseURL, l.config.GitHubHook.Port)

	serverErr := make(chan error, 1)
	go func() {
		logrus.Infof("Github webhook listener start httpServer on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("failed to start httpServer: %w", err)
		}
	}()

	select {
	case err := <-serverErr:
		logrus.Errorf("Github webhook httpServer has error: %v", err)
	case <-ctx.Done():
		logrus.Infof("Github webhook httpServer received shutdown signal")
	}

	return nil
}
