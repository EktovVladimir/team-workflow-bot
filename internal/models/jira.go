package models

import (
	"fmt"
	"regexp"
	"strings"

	"team-workflow-bot/internal/constants"
)

type IssueRef struct {
	Owner   string
	Project string
	Number  string
}

func (ref *IssueRef) ToUrl() string {
	owner := strings.TrimSpace(ref.Owner)
	if owner == "" {
		owner = constants.DefaultJiraOwner
	}
	return fmt.Sprintf("https://%s.atlassian.net/browse/%s", owner, ref.Number)
}

func ParseIssueRefFromUrl(url string) (*IssueRef, bool) {
	// Группы: 1) owner (label до .atlassian.net) 2) проект 3) номер
	jiraUrlRe := regexp.MustCompile(`^(?:https?://)?([A-Za-z0-9-]+)\.atlassian\.net/browse/([A-Za-z0-9]+)-(\d+)(?:/.*)?(?:[?#].*)?$`)
	m := jiraUrlRe.FindStringSubmatch(url)
	if len(m) != 4 {
		return nil, false
	}
	owner := strings.TrimSpace(m[1])
	project := m[2]
	number := m[3]

	if owner == "" {
		owner = constants.DefaultJiraOwner
	}

	fullKey := project + "-" + number
	return &IssueRef{Owner: owner, Project: project, Number: fullKey}, true
}

type IssueInfo struct {
	Ref   *IssueRef
	Title string
}
