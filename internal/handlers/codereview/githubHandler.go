package codereview

import (
	"context"
	"errors"
	"log"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/githubflow"
	"team-workflow-bot/internal/models"

	"github.com/cbrgm/githubevents/v2/githubevents"
	"github.com/google/go-github/v79/github"
)

func (h *Handler) HandlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent) {
	log.Printf("Handle PR review event: %s", event.GetAction())

	action := event.GetAction()

	prInfo := githubflow.MapPullRequestInfoFromResponse(event.PullRequest)

	if action == githubevents.PullRequestEventOpenedAction ||
		action == githubevents.PullRequestEventReadyForReviewAction ||
		action == githubevents.PullRequestEventUnlabeledAction {

		h.handlePullRequestReady(ctx, event, prInfo)
	} else if action == githubevents.PullRequestEventClosedAction && prInfo.IsMerged {

		h.handlePullRequestMerged(ctx, event, prInfo)
	}
}

func (h *Handler) HandlePullRequestReviewEvent(ctx context.Context, event *github.PullRequestReviewEvent) {
	log.Printf("Handle PR review event: %s", event.GetAction())
}

func (h *Handler) handlePullRequestReady(ctx context.Context, event *github.PullRequestEvent, prInfo *models.PullRequestInfo) {

}

func (h *Handler) handlePullRequestMerged(ctx context.Context, event *github.PullRequestEvent, prInfo *models.PullRequestInfo) {
	dbCrThread, err := h.bag.DB.Repository.GetCodeReviewThreadByPullRequestRef(ctx, prInfo.Ref)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return
		}

		log.Printf("Failed to get code review thread by PR ref: %v", err)
		return
	}
	//TODO
	_ = dbCrThread
}
