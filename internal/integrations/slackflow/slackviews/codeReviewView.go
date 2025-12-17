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
	CrPreviewEditModal = "cr_preview_edit_modal"

	CrPreviewEdit       = "cr_preview_edit"
	CrPreviewLoading    = "cr_preview_loading"
	CrThreadEdit        = "cr_thread_edit"
	CrThread            = "cr_thread"
	CrThreadContextMenu = "cr_thread_context_menu"

	ChannelField            = "channel"
	ReviewersField          = "reviewers"
	PullRequestRawListField = "pull_request_raw_list"
	IssueRawListField       = "issue_raw_list"

	CrEditCancel           = "cancel"
	CrEditConfirmAndPost   = "confirm_post"
	CrEditConfirmAndUpdate = "confirm_update"
	CrEditDelete           = "delete"

	CrThreadShowContextMenu = "show_context_menu"
	CrShowDeleteConfirm     = "show_delete_confirm"
	CrShowUpdateForm        = "show_update_form"
)

func GetCrPreviewEditBlocks(cr *models.CodeReviewContext, base string, resetIssueList bool) []slack.Block {
	validReviewers, notFoundGhReviewers := resolveReviewerNames(cr.Reviewers)

	res := []slack.Block{
		slackutils.GetUserMultiSelectInputBlock(
			slackutils.GetBlockId(base, ReviewersField),
			slackutils.GetActionId(base, ReviewersField),
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
			slackutils.GetBlockId(base, PullRequestRawListField),
			slackutils.GetActionId(base, PullRequestRawListField),
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

	issueListBlockId := slackutils.GetBlockId(base, IssueRawListField)
	if resetIssueList {
		issueListBlockId = slackutils.GetBlockIdWithRandSuffix(base, IssueRawListField)
	}

	res = append(res,
		slackutils.GetTextInputBlock(
			issueListBlockId,
			slackutils.GetActionId(base, IssueRawListField),
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

func GetCrThreadInternalInfoBlocks() []slack.Block {
	res := []slack.Block{
		slackutils.GetMarkdownTextSectionBlock("Зарезервировано для дополнительной информации"),
		slack.NewSectionBlock(
			slackutils.GetMarkdownTextObject("Показать доступные действия:"),
			nil,
			slack.NewAccessory(slack.NewButtonBlockElement(
				slackutils.GetActionId(CrThread, CrThreadShowContextMenu),
				CrThreadShowContextMenu,
				slackutils.GetPlainTextObject(":arrow_down_small:")))),
	}

	return res
}

func GetCrPreviewEditAndConfirmBlocks(cr *models.CodeReviewContext, resetIssueList bool) []slack.Block {
	base := CrPreviewEdit

	inputs := GetCrPreviewEditBlocks(cr, base, resetIssueList)

	cancelButton := slack.NewButtonBlockElement(
		slackutils.GetActionId(base, CrEditCancel),
		slackutils.GetActionId(base, CrEditCancel),
		slackutils.GetEmojiPlainTextObject("Отмена :x:"))

	confirmButton := slack.NewButtonBlockElement(
		slackutils.GetActionId(base, CrEditConfirmAndPost),
		slackutils.GetActionId(base, CrEditConfirmAndPost),
		slackutils.GetEmojiPlainTextObject("Создать тред :check_mark: "))

	confirmActionBlock := slack.NewActionBlock(
		slackutils.GetBlockId(base, CrEditConfirmAndPost),
		cancelButton,
		confirmButton,
	)

	res := append(inputs, confirmActionBlock)

	return res
}

func GetCrThreadUpdateBlocks(cr *models.CodeReviewContext) []slack.Block {
	base := CrThreadEdit

	inputs := GetCrPreviewEditBlocks(cr, base, false)

	cancelButton := slack.NewButtonBlockElement(
		slackutils.GetActionId(base, CrEditCancel),
		CrEditCancel,
		slackutils.GetEmojiPlainTextObject("Отмена :x:"))

	confirmButton := slack.NewButtonBlockElement(
		slackutils.GetActionId(base, CrEditConfirmAndUpdate),
		CrEditConfirmAndUpdate,
		slackutils.GetEmojiPlainTextObject("Обновить тред :check_mark:"))

	confirmActionBlock := slack.NewActionBlock(
		slackutils.GetBlockId(base, CrEditConfirmAndUpdate),
		cancelButton,
		confirmButton,
	)

	res := append(inputs, confirmActionBlock)

	return res
}

func GetCrPreviewModal(cr *models.CodeReviewContext, channelId string, resetIssueList bool) slack.ModalViewRequest {
	base := CrPreviewEdit

	blocks := make([]slack.Block, 0)
	blocks = append(blocks, slackutils.GetChannelInputBlock(
		slackutils.GetBlockId(base, ChannelField),
		slackutils.GetActionId(base, ChannelField),
		"Канал для создания треда",
		slackutils.WithInitialValue(channelId)))
	blocks = append(blocks, GetCrPreviewEditBlocks(cr, base, resetIssueList)...)

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

func GetCrPreviewLoadingModal() slack.ModalViewRequest {
	blocks := []slack.Block{
		slackutils.GetMarkdownTextSectionBlock(":loading1: Секундочку, работаем..."),
	}
	blocks = append(blocks, GetCrPreviewEditBlocks(&models.CodeReviewContext{}, CrPreviewLoading, false)...)

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

func GetCrThreadContextMenuBlocks() []slack.Block {
	base := CrThreadContextMenu

	updateBtn := slack.NewButtonBlockElement(
		slackutils.GetActionId(base, CrShowUpdateForm),
		slackutils.GetActionId(base, CrShowUpdateForm),
		slackutils.GetEmojiPlainTextObject("Редактировать"))

	deleteBtn := slack.NewButtonBlockElement(
		slackutils.GetActionId(base, CrShowDeleteConfirm),
		slackutils.GetActionId(base, CrShowDeleteConfirm),
		slackutils.GetEmojiPlainTextObject("Удалить #CR тред"))
	deleteBtn.Style = slack.StyleDanger

	return []slack.Block{
		slack.NewActionBlock(
			//TODO blockId пока не имеет значения.
			slackutils.GetBlockId(base, CrShowUpdateForm),
			updateBtn,
			deleteBtn,
		),
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

func GetReviewersWithActionButtonBlocks(reviewers []*models.UserRef) []slack.Block {
	validNames, notFoundGithubLogins := resolveReviewerNames(reviewers)

	reviewersText := strings.Join(lo.Map(validNames, func(s string, _ int) string {
		return fmt.Sprintf("<@%s>", s)
	}), " ")

	res := []slack.Block{
		slack.NewSectionBlock(
			slackutils.GetMarkdownTextObject(fmt.Sprintf("*#CR* %s", reviewersText)),
			nil,
			slack.NewAccessory(slack.NewButtonBlockElement(
				slackutils.GetActionId(CrThread, CrThreadShowContextMenu),
				CrThreadShowContextMenu,
				slackutils.GetPlainTextObject(":pencil:")))),
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
