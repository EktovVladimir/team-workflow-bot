package models

import "github.com/samber/lo"

type CodeReviewCollectRequest struct {
	Requester    *UserRef
	Reviewers    []*UserRef
	PullRequests []*PullRequestRef
	Issues       []*IssueRef

	DisableCollectReviewersFromPr bool
	DisableCollectIssuesFromPr    bool
}

func (r *CodeReviewCollectRequest) GetIssueKeys() []string {
	res := lo.Map(r.Issues, func(issueRef *IssueRef, _ int) string {
		return issueRef.Number
	})
	return lo.Uniq(res)
}

type CodeReviewContext struct {
	Key          string
	Requester    *UserRef
	Reviewers    []*UserRef
	PullRequests []*PullRequestInfo
	Issues       []*IssueInfo
}
