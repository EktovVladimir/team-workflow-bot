package bag

import (
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/githubflow"
	"team-workflow-bot/internal/integrations/jiraflow"
	"team-workflow-bot/internal/integrations/slackflow"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v79/github"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/mongo"
)

type DependenciesBag struct {
	DB       DatabaseBag
	Client   ClientBag
	Services InfrastructureServiceBag
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

type InfrastructureServiceBag struct {
	Github *githubflow.Service
	Jira   *jiraflow.Service
	Slack  *slackflow.Service
}

type ServiceWithDependencies struct {
	Bag *DependenciesBag
}

func NewServiceWithDependencies(bag *DependenciesBag) *ServiceWithDependencies {
	return &ServiceWithDependencies{
		Bag: bag,
	}
}
