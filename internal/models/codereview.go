package models

import (
	"github.com/samber/lo"
)

const (
	CodeReviewStatusOpen       = "open"
	CodeReviewStatusHalfMerged = "half_merged"
	CodeReviewStatusMerged     = "merged"
	CodeReviewStatusCanceled   = "canceled"
	CodeReviewStatusDeployed   = "deployed"
)

type CodeReviewThread struct {
	Id               UniqId             `bson:"_id,omitempty"`
	Thread           *MessageRef        `bson:"thread"`
	InternalMessages []*MessageRef      `bson:"internal_messages"`
	MessageLink      string             `bson:"message_link,omitempty"`
	Status           string             `bson:"status,omitempty"`
	Context          *CodeReviewContext `bson:"context"`

	SoftDeletable `bson:",inline"`
	Auditable     `bson:",inline"`
}

type CodeReviewContext struct {
	KeyedIssue   string             `bson:"keyed_issue"`
	Requester    *UserRef           `bson:"requester"`
	Reviewers    []*UserRef         `bson:"reviewers"`
	PullRequests []*PullRequestInfo `bson:"pull_requests"`
	Issues       []*IssueInfo       `bson:"issues"`
}

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
