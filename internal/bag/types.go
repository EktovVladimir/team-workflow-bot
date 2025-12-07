package bag

import (
	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v79/github"
	"github.com/slack-go/slack"
)

type DependenciesBag struct {
	Client ClientsBag
}

type ClientsBag struct {
	GitHub *github.Client
	Slack  *slack.Client
	Jira   *jira.Client
}
