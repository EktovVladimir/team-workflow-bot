package codereview

import (
	"team-workflow-bot/internal/models"

	"github.com/samber/lo"
)

type CrPreviewFormData struct {
	RequesterSlackId string
	ReviewerSlackIds []string
	PullRequestUrls  []string
	IssueUrls        []string

	ChannelId string
	AsUser    bool
}

func (f *CrPreviewFormData) ToCodeReviewCollectRequest() *models.CodeReviewCollectRequest {
	requester := &models.UserRef{
		SlackId: f.RequesterSlackId,
	}

	reviewerRefs := lo.Map(f.ReviewerSlackIds, func(id string, i int) *models.UserRef {
		return &models.UserRef{
			SlackId: id,
		}
	})

	prRefs := getParsedFromUrlPullRequestRefs(f.PullRequestUrls)
	issueRefs := getParsedFromUrlIssueRefs(f.IssueUrls)

	return &models.CodeReviewCollectRequest{
		Requester:    requester,
		Reviewers:    reviewerRefs,
		PullRequests: prRefs,
		Issues:       issueRefs,
	}
}
