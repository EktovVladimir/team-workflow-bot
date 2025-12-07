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
	config *config.Config
}

func NewHandler(config *config.Config) *Handler {
	return &Handler{
		config: config,
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
	//TODO
	log.Printf("Handling GitHub PR event %v: %+v", eventName, event)
	return nil
}

func (l *Handler) handlePullRequestReviewEvent(
	ctx context.Context,
	deliveryID string,
	eventName string,
	event *github.PullRequestReviewEvent) error {
	//TODO
	log.Printf("Handling GitHub PR review event %v: %+v", eventName, event)
	return nil
}

func (l *Handler) handleWorkflowRunEvent(
	ctx context.Context,
	deliveryID string,
	eventName string,
	event *github.WorkflowRunEvent) error {
	//TODO
	log.Printf("Handling GitHub WF run event %v: %+v", eventName, event)
	return nil
}
