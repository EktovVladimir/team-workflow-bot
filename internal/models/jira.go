package models

type IssueRef struct {
	Owner   string
	Project string
	Number  string
}

type IssueInfo struct {
	Ref   *IssueRef
	Title string
}
