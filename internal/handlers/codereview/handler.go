package codereview

import (
	"context"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/models"
)

// Золотистый
type retriever interface {
	CollectCodeReviewContextFromSlack(ctx context.Context, request *models.CodeReviewCollectRequest) (*models.CodeReviewContext, error)
}

type Handler struct {
	*bag.ServiceWithDependencies

	retriever retriever

	slackService *slackflow.Service
	repository   *db.Repository
}

func newHandler(b *bag.DependenciesBag, retriever retriever) *Handler {
	return &Handler{
		ServiceWithDependencies: bag.NewServiceWithDependencies(b),
		retriever:               retriever,
		slackService:            b.Services.Slack,
		repository:              b.DB.Repository,
	}
}
