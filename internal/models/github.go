package models

type PullRequestRef struct {
	Owner  string
	Repo   string
	Number int
}

type PullRequestInfo struct {
	Ref        PullRequestRef
	HeadBranch string
	BaseBranch string
	Title      string
	Requester  string
	Reviewers  []string
	Labels     []string
	State      string
	IsDraft    bool
	IsMerged   bool
}

type CommitInfo struct {
	Message string
}
