package githubflow

import (
	"context"
	"log"
	"net/http"
	"team-workflow-bot/internal/config"

	"github.com/cbrgm/githubevents/v2/githubevents"
	"github.com/google/go-github/v79/github"
)

type HttpHandler func(w http.ResponseWriter, r *http.Request)

type Handler struct {
	config        *config.Config
	optionsConfig *handlerOptionConfig
}

func NewHandler(config *config.Config, options ...HandlerOption) *Handler {
	optionsConfig := newHandlerOptionConfig(options...)

	return &Handler{
		config:        config,
		optionsConfig: optionsConfig,
	}
}

func (l *Handler) GetHttpHandler() HttpHandler {
	secretKey := l.config.GitHubHook.SecretKey

	handle := githubevents.New(secretKey)

	handle.OnPullRequestEventAny(l.handlePullRequestEvent)
	handle.OnPullRequestReviewEventAny(l.handlePullRequestReviewEvent)
	handle.OnWorkflowRunEventAny(l.handleWorkflowRunEvent)

	return func(w http.ResponseWriter, r *http.Request) {
		err := handle.HandleEventRequest(r)
		if err != nil {
			log.Printf("Error handling GitHub event: %v", err)
		}
	}
}

func (l *Handler) handlePullRequestEvent(
	ctx context.Context,
	deliveryID string,
	eventName string, event *github.PullRequestEvent) error {
	for _, h := range l.optionsConfig.prHandlers {
		h.HandlePullRequestEvent(ctx, event)
	}

	return nil
}

func (l *Handler) handlePullRequestReviewEvent(
	ctx context.Context,
	deliveryID string,
	eventName string,
	event *github.PullRequestReviewEvent) error {
	for _, h := range l.optionsConfig.prReviewHandlers {
		h.HandlePullRequestReviewEvent(ctx, event)
	}

	return nil
}

func (l *Handler) handleWorkflowRunEvent(
	ctx context.Context,
	deliveryID string,
	eventName string,
	event *github.WorkflowRunEvent) error {
	for _, h := range l.optionsConfig.wfRunHandler {
		h.HandleWorkflowRunEvent(ctx, event)
	}

	return nil
}
