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
		return issueRef.Key
	})
	return lo.Uniq(res)
}

type CodeReviewContext struct {
	KeyedIssue   string             `bson:"keyed_issue"`
	Requester    *UserRef           `bson:"requester"`
	Reviewers    []*UserRef         `bson:"reviewers"`
	PullRequests []*PullRequestInfo `bson:"pull_requests"`
	Issues       []*IssueInfo       `bson:"issues"`
}
