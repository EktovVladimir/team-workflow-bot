package models

type UserRef struct {
	SlackId     string `bson:"slack_id"`
	GithubLogin string `bson:"github_login"`
}

func (u *UserRef) IsEmpty() bool {
	return u.SlackId == "" && u.GithubLogin == ""
}
