package models

const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleUser      = "user"
)

type User struct {
	Id          string   `bson:"_id,omitempty"`
	Email       string   `bson:"email"`
	SlackName   string   `bson:"slack_name"`
	SlackId     string   `bson:"slack_id"`
	GitHubLogin string   `bson:"github_login"`
	Roles       []string `bson:"roles,omitempty"`
	Teams       []string `bson:"teams,omitempty"`
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
	Id          string `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Description string `bson:"description,omitempty"`
}

type Team struct {
	Id          string `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Description string `bson:"description,omitempty"`
	IssueRegex  string `bson:"issue_regex,omitempty"`
	Channel     string `bson:"channel,omitempty"`
}
