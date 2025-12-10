package models

type RequestRef struct {
	Requester    UserRef
	Reviewers    []UserRef
	PullRequests []PullRequestRef
	Issues       []IssueRef
}

type CodeReviewContext struct {
	Requester    *UserRef
	Reviewers    []*UserRef
	PullRequests []*PullRequestInfo
	Issues       []*IssueInfo
}
