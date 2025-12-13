package slackviews

import (
	"fmt"
	"strings"
	"team-workflow-bot/internal/models"

	"github.com/samber/lo"
	"github.com/slack-go/slack"
)

const (
	CrPreviewEditModal      = "cr_preview_edit_modal"
	CrPreviewEdit           = "cr_preview_edit"
	ChannelField            = "channel"
	ReviewersField          = "reviewers"
	PullRequestRawListField = "pull_request_raw_list"
	IssueRawListField       = "issue_raw_list"

	CrPreviewEditCancel  = "cancel"
	CrPreviewEditConfirm = "confirm"
)

func GetCrPreviewEditBlocks(cr *models.CodeReviewContext, resetIssueList bool) []slack.Block {
	validReviewers, notFoundGhReviewers := resolveReviewerNames(cr.Reviewers)

	res := []slack.Block{
		GetUserMultiSelectInputBlock(
			GetBlockId(CrPreviewEdit, ReviewersField),
			GetActionId(CrPreviewEdit, ReviewersField),
			"Reviewers",
			WithInitialValues(validReviewers)),
	}

	notFoundReviewersBlock := getNotValidReviewersContextBlock(notFoundGhReviewers)
	if notFoundReviewersBlock != nil {
		res = append(res, notFoundReviewersBlock)
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
			WithDispatchAction(true),
			WithInitialValue(prListInputText),
			WithMultiline(true),
			WithPlaceholder("Список PR'ов. По одному в строке."),
			WithHint("Можно редактировать список. По одной ссылке на строку. В итоговый тред попадут указанные здесь PR'ы.")))
	res = append(res, GetPullRequestListBlocks(cr.PullRequests...)...)

	res = append(res, slack.NewDividerBlock())

	issueUrls := lo.Map(cr.Issues, func(issue *models.IssueInfo, _ int) string {
		return issue.Ref.ToUrl()
	})
	issuesListInputText := strings.Join(issueUrls, "\n")

	issueListBlockId := GetBlockId(CrPreviewEdit, IssueRawListField)
	if resetIssueList {
		issueListBlockId = GetBlockIdWithRandSuffix(CrPreviewEdit, IssueRawListField)
	}

	res = append(res,
		GetTextInputBlock(
			issueListBlockId,
			GetActionId(CrPreviewEdit, IssueRawListField),
			"Задачи",
			WithDispatchAction(true),
			WithInitialValue(issuesListInputText),
			WithMultiline(true),
			WithOptional(true),
			WithPlaceholder("Список URL задач. По одному в строке."),
			WithHint("Можно редактировать список. По одной ссылке на строку. В итоговый тред попадут указанные здесь задачи.")))
	res = append(res, GetIssueListBlocks(cr.Issues...)...)

	return res
}

func GetCrThreadBlocks(cr *models.CodeReviewContext) []slack.Block {
	res := make([]slack.Block, 0)

	res = append(res, GetReviewersBlocks(cr.Reviewers)...)
	res = append(res, slack.NewDividerBlock())

	if len(cr.Issues) > 0 {
		res = append(res, GetMarkdownTextSectionBlock("*Issues:*"))
		res = append(res, GetIssueListBlocks(cr.Issues...)...)
		res = append(res, slack.NewDividerBlock())
	}

	res = append(res, GetMarkdownTextSectionBlock("*Pull requests:*"))
	res = append(res, GetPullRequestListBlocks(cr.PullRequests...)...)
	res = append(res, slack.NewDividerBlock())

	requesterSlackName := cr.Requester.SlackId
	contextText := fmt.Sprintf("Создано через бота по запросу <@%s>", requesterSlackName)

	res = append(res, GetSimpleMarkdownContextBlock(contextText))

	return res
}

func GetCrPreviewEditAndConfirmBlocks(cr *models.CodeReviewContext, resetIssueList bool) []slack.Block {
	inputs := GetCrPreviewEditBlocks(cr, resetIssueList)

	cancelButton := slack.NewButtonBlockElement(
		GetActionId(CrPreviewEdit, CrPreviewEditCancel),
		GetActionId(CrPreviewEdit, CrPreviewEditCancel),
		GetEmojiPlainTextObject("Отмена :x:"))

	confirmButton := slack.NewButtonBlockElement(
		GetActionId(CrPreviewEdit, CrPreviewEditConfirm),
		GetActionId(CrPreviewEdit, CrPreviewEditConfirm),
		GetEmojiPlainTextObject("Создать тред :check_mark: "))

	confirmActionBlock := slack.NewActionBlock(
		GetBlockId(CrPreviewEdit, CrPreviewEditConfirm),
		cancelButton,
		confirmButton,
	)

	res := append(inputs, confirmActionBlock)

	return res
}

func GetCrPreviewModal(cr *models.CodeReviewContext, channelId string, resetIssueList bool) slack.ModalViewRequest {

	blocks := make([]slack.Block, 0)
	blocks = append(blocks, GetChannelInputBlock(
		GetBlockId(CrPreviewEdit, ChannelField),
		GetActionId(CrPreviewEdit, ChannelField),
		"Канал для создания треда",
		WithInitialValue(channelId)))
	blocks = append(blocks, GetCrPreviewEditBlocks(cr, resetIssueList)...)

	return slack.ModalViewRequest{
		CallbackID: GetCallbackId(CrPreviewEditModal),
		Type:       slack.VTModal,
		Title:      GetEmojiPlainTextObject("Создание треда #CR"),
		Submit:     GetSimplePlainTextObject("Подтвердить"),
		Close:      GetSimplePlainTextObject("Отмена"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
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
			":jira: <%s|%s - %s>",
			issue.Ref.ToUrl(),
			issue.Ref.Key,
			issue.Title)

		res = append(res, GetMarkdownTextSectionBlock(text))
	}

	return res
}

func GetReviewersBlocks(reviewers []*models.UserRef) []slack.Block {
	validNames, notFoundGithubLogins := resolveReviewerNames(reviewers)

	reviewersText := strings.Join(lo.Map(validNames, func(s string, _ int) string {
		return fmt.Sprintf("<@%s>", s)
	}), " ")

	res := []slack.Block{
		GetMarkdownTextSectionBlock(fmt.Sprintf("*#CR* %s", reviewersText)),
	}

	notValidBlock := getNotValidReviewersContextBlock(notFoundGithubLogins)
	if notValidBlock != nil {
		res = append(res, notValidBlock)
	}

	return res
}

func getNotValidReviewersContextBlock(notFoundGithubLogins []string) slack.Block {
	if len(notFoundGithubLogins) == 0 {
		return nil
	}

	text := fmt.Sprintf(
		":warning: Некоторые ревьюверы не были найдены по их GitHub лоигнам: %s",
		wrapCodeQuotesAndJoin(notFoundGithubLogins...))

	return GetSimpleMarkdownContextBlock(text)
}

func resolveReviewerNames(reviewers []*models.UserRef) ([]string, []string) {
	validReviewers := make([]string, 0)
	notFoundGhReviewers := make([]string, 0)
	for _, reviewer := range reviewers {
		if reviewer.SlackId != "" {
			validReviewers = append(validReviewers, reviewer.SlackId)
		} else if reviewer.GithubLogin != "" {
			notFoundGhReviewers = append(notFoundGhReviewers, reviewer.GithubLogin)
		}
	}
	return validReviewers, notFoundGhReviewers
}

func wrapCodeQuotesAndJoin(items ...string) string {
	return strings.Join(lo.Map(items, func(s string, _ int) string {
		return fmt.Sprintf("`%s`", s)
	}), ", ")
}
