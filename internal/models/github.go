package models

import (
	"fmt"
	"regexp"
	"strconv"
)

type PullRequestRef struct {
	Owner  string
	Repo   string
	Number int
}

func (ref *PullRequestRef) ToUrl() string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", ref.Owner, ref.Repo, ref.Number)
}

func ParsePullRequestRefFromUrl(url string) (*PullRequestRef, bool) {
	prUrlRe := regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/pull/(\d+)(?:/.*)?(?:[?#].*)?$`)
	m := prUrlRe.FindStringSubmatch(url)
	if len(m) != 4 {
		return nil, false
	}
	owner := m[1]
	repo := m[2]
	// Преобразуем номер PR в int
	num, err := strconv.Atoi(m[3])
	if err != nil {
		return nil, false
	}
	return &PullRequestRef{Owner: owner, Repo: repo, Number: num}, true
}

type PullRequestInfo struct {
	Ref        *PullRequestRef
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
