package bag

import (
	"team-workflow-bot/internal/db"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v79/github"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/mongo"
)

type DependenciesBag struct {
	Client ClientBag
	DB     DatabaseBag
}

type ClientBag struct {
	GitHub *github.Client
	Slack  *slack.Client
	Jira   *jira.Client
}

type DatabaseBag struct {
	Mongo      *mongo.Database
	Repository *db.Repository
}
