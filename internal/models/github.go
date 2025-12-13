package models

import (
	"fmt"
	"regexp"
	"strconv"
)

type PullRequestRef struct {
	Owner  string `bson:"owner"`
	Repo   string `bson:"repo"`
	Number int    `bson:"number"`
}

func (ref *PullRequestRef) ToUrl() string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", ref.Owner, ref.Repo, ref.Number)
}

func (ref *PullRequestRef) ToKey() string {
	return fmt.Sprintf("%s/%s#%d", ref.Owner, ref.Repo, ref.Number)
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
	Ref        *PullRequestRef `bson:"ref"`
	RefKey     string          `bson:"ref_key"`
	HeadBranch string          `bson:"head_branch"`
	BaseBranch string          `bson:"base_branch"`
	Title      string          `bson:"title"`
	Requester  string          `bson:"requester"`
	Reviewers  []string        `bson:"reviewers"`
	Labels     []string        `bson:"labels"`
	State      string          `bson:"state"`
	IsDraft    bool            `bson:"is_draft"`
	IsMerged   bool            `bson:"is_merged"`
}

type CommitInfo struct {
	Message string
}
