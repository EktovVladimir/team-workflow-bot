package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	//TODO более гибкие роли и права
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleUser      = "user"
)

const (
	CodeReviewStatusOpen     = "open"
	CodeReviewStatusClosed   = "closed"
	CodeReviewStatusCanceled = "canceled"
)

type UniqId = primitive.ObjectID

type Auditable struct {
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	CreatedBy UniqId    `bson:"created_by"`
	UpdatedBy UniqId    `bson:"updated_by"`
}

type User struct {
	Id              UniqId   `bson:"_id,omitempty"`
	Email           string   `bson:"email"`
	SlackName       string   `bson:"slack_name"`
	SlackId         string   `bson:"slack_id"`
	GitHubLogin     string   `bson:"github_login"`
	AlternateLogins []string `bson:"alternate_logins,omitempty"`
	Roles           []string `bson:"roles,omitempty"`
	Teams           []string `bson:"teams,omitempty"`
}

func (u *User) ToRef() *UserRef {
	return &UserRef{
		SlackId:     u.SlackId,
		GithubLogin: u.GitHubLogin,
	}
}

func (u *User) HasRole(roleName string) bool {
	for _, r := range u.Roles {
		if r == roleName {
			return true
		}
	}
	return false
}

type Role struct {
	Id          UniqId `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Description string `bson:"description,omitempty"`
}

type Team struct {
	Id          UniqId `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Description string `bson:"description,omitempty"`
	IssueRegex  string `bson:"issue_regex,omitempty"`
	Channel     string `bson:"channel,omitempty"`
}

type CodeReviewThread struct {
	Id          UniqId             `bson:"_id,omitempty"`
	Thread      *ThreadRef         `bson:"thread"`
	MessageLink string             `bson:"message_link,omitempty"`
	Status      string             `bson:"status,omitempty"`
	Context     *CodeReviewContext `bson:"context"`

	Auditable `bson:",inline"`
}
