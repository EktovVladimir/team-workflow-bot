package slackviews

import (
	"fmt"
	"strings"
	"team-workflow-bot/internal/models"
	"team-workflow-bot/pkg/slackutils"

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
		slackutils.GetUserMultiSelectInputBlock(
			slackutils.GetBlockId(CrPreviewEdit, ReviewersField),
			slackutils.GetActionId(CrPreviewEdit, ReviewersField),
			"Reviewers",
			slackutils.WithInitialValues(validReviewers)),
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
		slackutils.GetTextInputBlock(
			slackutils.GetBlockId(CrPreviewEdit, PullRequestRawListField),
			slackutils.GetActionId(CrPreviewEdit, PullRequestRawListField),
			"Pull requests",
			slackutils.WithDispatchAction(true),
			slackutils.WithInitialValue(prListInputText),
			slackutils.WithMultiline(true),
			slackutils.WithPlaceholder("Список PR'ов. По одному в строке."),
			slackutils.WithHint("Можно редактировать список. По одной ссылке на строку. В итоговый тред попадут указанные здесь PR'ы.")))
	res = append(res, GetPullRequestListBlocks(cr.PullRequests...)...)

	res = append(res, slack.NewDividerBlock())

	issueUrls := lo.Map(cr.Issues, func(issue *models.IssueInfo, _ int) string {
		return issue.Ref.ToUrl()
	})
	issuesListInputText := strings.Join(issueUrls, "\n")

	issueListBlockId := slackutils.GetBlockId(CrPreviewEdit, IssueRawListField)
	if resetIssueList {
		issueListBlockId = slackutils.GetBlockIdWithRandSuffix(CrPreviewEdit, IssueRawListField)
	}

	res = append(res,
		slackutils.GetTextInputBlock(
			issueListBlockId,
			slackutils.GetActionId(CrPreviewEdit, IssueRawListField),
			"Задачи",
			slackutils.WithDispatchAction(true),
			slackutils.WithInitialValue(issuesListInputText),
			slackutils.WithMultiline(true),
			slackutils.WithOptional(true),
			slackutils.WithPlaceholder("Список URL задач. По одному в строке."),
			slackutils.WithHint("Можно редактировать список. По одной ссылке на строку. В итоговый тред попадут указанные здесь задачи.")))
	res = append(res, GetIssueListBlocks(cr.Issues...)...)

	return res
}

func GetCrThreadBlocks(cr *models.CodeReviewContext) []slack.Block {
	res := make([]slack.Block, 0)

	res = append(res, GetReviewersBlocks(cr.Reviewers)...)
	res = append(res, slack.NewDividerBlock())

	if len(cr.Issues) > 0 {
		res = append(res, slackutils.GetMarkdownTextSectionBlock("*Issues:*"))
		res = append(res, GetIssueListBlocks(cr.Issues...)...)
		res = append(res, slack.NewDividerBlock())
	}

	res = append(res, slackutils.GetMarkdownTextSectionBlock("*Pull requests:*"))
	res = append(res, GetPullRequestListBlocks(cr.PullRequests...)...)
	res = append(res, slack.NewDividerBlock())

	requesterSlackName := cr.Requester.SlackId
	contextText := fmt.Sprintf("Создано через бота по запросу <@%s>", requesterSlackName)

	res = append(res, slackutils.GetSimpleMarkdownContextBlock(contextText))

	return res
}

func GetCrPreviewEditAndConfirmBlocks(cr *models.CodeReviewContext, resetIssueList bool) []slack.Block {
	inputs := GetCrPreviewEditBlocks(cr, resetIssueList)

	cancelButton := slack.NewButtonBlockElement(
		slackutils.GetActionId(CrPreviewEdit, CrPreviewEditCancel),
		slackutils.GetActionId(CrPreviewEdit, CrPreviewEditCancel),
		slackutils.GetEmojiPlainTextObject("Отмена :x:"))

	confirmButton := slack.NewButtonBlockElement(
		slackutils.GetActionId(CrPreviewEdit, CrPreviewEditConfirm),
		slackutils.GetActionId(CrPreviewEdit, CrPreviewEditConfirm),
		slackutils.GetEmojiPlainTextObject("Создать тред :check_mark: "))

	confirmActionBlock := slack.NewActionBlock(
		slackutils.GetBlockId(CrPreviewEdit, CrPreviewEditConfirm),
		cancelButton,
		confirmButton,
	)

	res := append(inputs, confirmActionBlock)

	return res
}

func GetCrPreviewModal(cr *models.CodeReviewContext, channelId string, resetIssueList bool) slack.ModalViewRequest {

	blocks := make([]slack.Block, 0)
	blocks = append(blocks, slackutils.GetChannelInputBlock(
		slackutils.GetBlockId(CrPreviewEdit, ChannelField),
		slackutils.GetActionId(CrPreviewEdit, ChannelField),
		"Канал для создания треда",
		slackutils.WithInitialValue(channelId)))
	blocks = append(blocks, GetCrPreviewEditBlocks(cr, resetIssueList)...)

	return slack.ModalViewRequest{
		CallbackID: slackutils.GetCallbackId(CrPreviewEditModal),
		Type:       slack.VTModal,
		Title:      slackutils.GetEmojiPlainTextObject("Создание треда #CR"),
		Submit:     slackutils.GetSimplePlainTextObject("Подтвердить"),
		Close:      slackutils.GetSimplePlainTextObject("Отмена"),
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

		res = append(res, slackutils.GetMarkdownTextSectionBlock(text))
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

		res = append(res, slackutils.GetMarkdownTextSectionBlock(text))
	}

	return res
}

func GetReviewersBlocks(reviewers []*models.UserRef) []slack.Block {
	validNames, notFoundGithubLogins := resolveReviewerNames(reviewers)

	reviewersText := strings.Join(lo.Map(validNames, func(s string, _ int) string {
		return fmt.Sprintf("<@%s>", s)
	}), " ")

	res := []slack.Block{
		slackutils.GetMarkdownTextSectionBlock(fmt.Sprintf("*#CR* %s", reviewersText)),
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

	return slackutils.GetSimpleMarkdownContextBlock(text)
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
