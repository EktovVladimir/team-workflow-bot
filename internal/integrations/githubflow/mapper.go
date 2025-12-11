package githubflow

import (
	"team-workflow-bot/internal/models"

	"github.com/google/go-github/v79/github"
	"github.com/samber/lo"
)

func MapPullRequestInfoFromResponse(ghPr *github.PullRequest) *models.PullRequestInfo {
	reviewers := lo.Map(ghPr.RequestedReviewers, func(u *github.User, i int) string {
		return u.GetLogin()
	})

	labels := lo.Map(ghPr.Labels, func(l *github.Label, i int) string {
		return l.GetName()
	})

	return &models.PullRequestInfo{
		Ref: &models.PullRequestRef{
			Owner:  ghPr.Base.Repo.Owner.GetLogin(),
			Repo:   ghPr.Base.Repo.GetName(),
			Number: ghPr.GetNumber(),
		},
		HeadBranch: ghPr.Head.GetRef(),
		BaseBranch: ghPr.Base.GetRef(),
		Title:      ghPr.GetTitle(),
		Requester:  ghPr.User.GetLogin(),
		State:      ghPr.GetState(),
		IsDraft:    ghPr.GetDraft(),
		IsMerged:   ghPr.GetMerged(),
		Reviewers:  reviewers,
		Labels:     labels,
	}
}
