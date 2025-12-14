package codereview

import (
	"context"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/models"
)

// Золотистый
type retriever interface {
	CollectCodeReviewContextFromSlack(ctx context.Context, request *models.CodeReviewCollectRequest) (*models.CodeReviewContext, error)
}

type Handler struct {
	bag       *bag.DependenciesBag
	retriever retriever
}

func NewHandler(bag *bag.DependenciesBag, retriever retriever) *Handler {
	return &Handler{
		bag:       bag,
		retriever: retriever,
	}
}
