package models

type UserRef struct {
	SlackId     string
	GithubLogin string
}

func (u *UserRef) IsEmpty() bool {
	return u.SlackId == "" && u.GithubLogin == ""
}
