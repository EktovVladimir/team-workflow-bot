package slackviews

import (
	"fmt"
	"strings"
	"team-workflow-bot/internal/models"

	"github.com/samber/lo"
	"github.com/slack-go/slack"
)

const (
	CrPreviewEdit           = "cr_preview_edit"
	ReviewersField          = "reviewers"
	PullRequestRawListField = "pull_request_raw_list"
	IssueRawListField       = "issue_raw_list"

	CrPreviewEditConfirm = "confirm"
)

func GetCrPreviewEditBlocks(cr *models.CodeReviewContext) []slack.Block {
	validReviewers := make([]string, 0)
	notFoundGhReviewers := make([]string, 0)
	for _, reviewer := range cr.Reviewers {
		if reviewer.SlackId != "" {
			validReviewers = append(validReviewers, reviewer.SlackId)
		} else if reviewer.GithubLogin != "" {
			notFoundGhReviewers = append(notFoundGhReviewers, reviewer.GithubLogin)
		}
	}

	res := []slack.Block{
		GetUserMultiSelectInputBlock(
			GetBlockId(CrPreviewEdit, ReviewersField),
			GetActionId(CrPreviewEdit, ReviewersField),
			"Reviewers",
			WithInitialValues(validReviewers)),
	}

	if len(notFoundGhReviewers) > 0 {
		text := fmt.Sprintf(
			":warning: Некоторые ревьюверы не были найдены по их GitHub лоигнам: `%s`",
			strings.Join(notFoundGhReviewers, ", "))

		res = append(res, GetSimpleMarkdownContextBlock(text))
	}

	res = append(res, slack.NewDividerBlock())

	prUrls := lo.Map(cr.PullRequests, func(pr *models.PullRequestInfo, _ int) string {
		return pr.Ref.ToUrl()
	})
	prListInputText := strings.Join(prUrls, "\n")

	res = append(res,
		GetTextInputBlock(
			GetBlockId(CrPreviewEdit, PullRequestRawListField),
			GetActionId(CrPreviewEdit, PullRequestRawListField),
			"Pull requests",
			WithInitialValue(prListInputText),
			WithMultiline(true),
			WithPlaceholder("Список PR'ов. По одному в строке."),
			WithHint("Можно редактировать список. По одной ссылке на строку. В итоговый тред попадут указанные здесь PR'ы.")))
	res = append(res, GetMarkdownTextSectionBlock("*Превью PR'ов на код-ревью:* "))
	res = append(res, GetPullRequestListBlocks(cr.PullRequests...)...)

	res = append(res, slack.NewDividerBlock())

	issueUrls := lo.Map(cr.Issues, func(issue *models.IssueInfo, _ int) string {
		return issue.Ref.ToUrl()
	})
	issuesListInputText := strings.Join(issueUrls, "\n")

	res = append(res,
		GetTextInputBlock(
			GetBlockId(CrPreviewEdit, IssueRawListField),
			GetActionId(CrPreviewEdit, IssueRawListField),
			"Задачи",
			WithInitialValue(issuesListInputText),
			WithMultiline(true),
			WithPlaceholder("Список URL задач. По одному в строке."),
			WithHint("Можно редактировать список. По одной ссылке на строку. В итоговый тред попадут указанные здесь задачи.")))
	res = append(res, GetMarkdownTextSectionBlock("*Превью задач:* "))
	res = append(res, GetIssueListBlocks(cr.Issues...)...)

	return res
}

func GetCrPreviewEditAndConfirmBlocks(cr *models.CodeReviewContext) []slack.Block {
	inputs := GetCrPreviewEditBlocks(cr)

	confirmButton := slack.NewButtonBlockElement(
		GetActionId(CrPreviewEdit, CrPreviewEditConfirm),
		//TODO надо ли в value что-то осмысленное?
		GetActionId(CrPreviewEdit, CrPreviewEditConfirm),
		GetSimplePlainTextObject("Создать тред"),
	)
	confirmActionBlock := slack.NewActionBlock(
		GetBlockId(CrPreviewEdit, CrPreviewEditConfirm),
		confirmButton,
	)

	res := append(inputs, confirmActionBlock)

	return res
}

func GetPullRequestListBlocks(prs ...*models.PullRequestInfo) []slack.Block {
	res := make([]slack.Block, 0)

	for _, pr := range prs {
		text := fmt.Sprintf(
			":git-hub: <%s|%s> (%s)",
			pr.Ref.ToUrl(),
			pr.Title,
			pr.Ref.Repo)

		res = append(res, GetMarkdownTextSectionBlock(text))
	}

	return res
}

func GetIssueListBlocks(issues ...*models.IssueInfo) []slack.Block {
	res := make([]slack.Block, 0)

	for _, issue := range issues {
		text := fmt.Sprintf(
			":jira: <%s|%s>",
			issue.Ref.ToUrl(),
			issue.Title)

		res = append(res, GetMarkdownTextSectionBlock(text))
	}

	return res
}
