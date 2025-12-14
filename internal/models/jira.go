package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"team-workflow-bot/internal/constants"
)

type IssueRef struct {
	Owner   string `bson:"owner"`
	Project string `bson:"project"`
	Number  int    `bson:"number"`
	Key     string `bson:"key"`
}

func (ref *IssueRef) ToUrl() string {
	owner := strings.TrimSpace(ref.Owner)
	if owner == "" {
		owner = constants.DefaultJiraOwner
	}
	return fmt.Sprintf("https://%s.atlassian.net/browse/%s", owner, ref.Key)
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

	parsedNumber, _ := strconv.Atoi(number)

	fullKey := project + "-" + number
	return &IssueRef{
		Owner:   owner,
		Project: project,
		Key:     fullKey,
		Number:  parsedNumber,
	}, true
}

type IssueInfo struct {
	Ref   *IssueRef
	Title string
}
