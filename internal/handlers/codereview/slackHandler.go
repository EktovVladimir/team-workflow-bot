package codereview

import (
	"context"
	"fmt"
	"strings"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/internal/models"
	"team-workflow-bot/pkg/slackutils"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	crThreadCreatedMetaEventType = "cr_request_created"
	handlerNameMetaField         = "handler"
	contextMetaField             = "context"
	handlerNameMetaValue         = "codereview.Handler"
)

func (h *Handler) HandleSlackSlashCommand(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, cmd slack.SlashCommand) {
	if !h.isCommandApplicable(cmd) {
		return
	}

	//TODO конфиг
	const (
		sendAsModal = true
	)

	client.Ack(*evt.Request)

	args := strings.Fields(cmd.Text)

	prRefs := getParsedFromUrlPullRequestRefs(args)
	issueRefs := getParsedFromUrlIssueRefs(args)
	reviewerRefs := getParsedFromFormatUserRefs(args)

	requester := &models.UserRef{
		SlackId: cmd.UserID,
	}

	request := &models.CodeReviewCollectRequest{
		Requester:    requester,
		Reviewers:    reviewerRefs,
		PullRequests: prRefs,
		Issues:       issueRefs,
	}

	crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)

	if err != nil {
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
			fmt.Sprintf("Ошибка при сборе контекста код-ревью: %s", err.Error()))
		return
	}

	//TODO пытаемся сохранить в БД новых юзеров?

	if sendAsModal {
		triggerId := cmd.TriggerID

		_, err = client.OpenViewContext(
			ctx,
			triggerId,
			slackviews.GetCrPreviewModal(crContext, cmd.ChannelID, false))
		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
				fmt.Sprintf("Ошибка при открытии модального окна превью код-ревью: %s", err.Error()))
			return
		}
	} else {
		notificationText := fmt.Sprintf("Превью #CR треда %s", crContext.KeyedIssue)

		_, _, _, err = client.SendMessageContext(
			ctx,
			cmd.ChannelID,
			slack.MsgOptionText(notificationText, false),
			slack.MsgOptionPostEphemeral(cmd.UserID),
			slack.MsgOptionBlocks(slackviews.GetCrPreviewEditAndConfirmBlocks(crContext, false)...))

		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
				fmt.Sprintf("Ошибка при отправке превью код-ревью: %s", err.Error()))
			return
		}
	}
}

func (h *Handler) HandleSlackBlockAction(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback) {
	sl := h.bag.Client.Slack

	if len(callback.ActionCallback.BlockActions) == 0 {
		return
	}

	actionId := callback.ActionCallback.BlockActions[0].ActionID
	responseUrl := callback.ResponseURL
	channelId := callback.Channel.ID
	userId := callback.User.ID

	var (
		crPreviewSubmitActionId    = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.CrPreviewEditConfirm)
		crPreviewCancelActionId    = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.CrPreviewEditCancel)
		crPreviewPrListActionId    = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.PullRequestRawListField)
		crPreviewIssueListActionId = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.IssueRawListField)
	)

	if actionId == crPreviewSubmitActionId {
		client.Ack(*evt.Request)

		form := getRequestRefFromPreviewEdit(callback.BlockActionState.Values, userId)
		request := form.ToCodeReviewCollectRequest()

		request.DisableCollectReviewersFromPr = true
		request.DisableCollectIssuesFromPr = true

		err := h.handleAndSendCrThread(ctx, request, channelId, form.AsUser)
		if err != nil {
			logrus.Error("Failed to handle and send cr thread:", err)
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, channelId, callback.User.ID,
				fmt.Sprintf("Не удалось создать #CR тред: %s", err.Error()))
			return
		}

		_, _, _ = sl.PostMessage(channelId, slack.MsgOptionDeleteOriginal(responseUrl))

		return
	}

	if actionId == crPreviewCancelActionId {
		client.Ack(*evt.Request)
		if responseUrl != "" {
			_, _, _ = sl.PostMessage(channelId, slack.MsgOptionDeleteOriginal(responseUrl))
		}

		return
	}

	if actionId == crPreviewPrListActionId || actionId == crPreviewIssueListActionId {
		var values map[string]map[string]slack.BlockAction

		if callback.Container.Type == "view" {
			values = callback.View.State.Values
		} else {
			values = callback.BlockActionState.Values
		}

		form := getRequestRefFromPreviewEdit(values, userId)
		request := form.ToCodeReviewCollectRequest()

		request.DisableCollectReviewersFromPr = true

		//TODO false для crPreviewPrListActionId, но сначала разобраться с обновлением инпутов
		request.DisableCollectIssuesFromPr = actionId == crPreviewIssueListActionId

		crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)
		if err != nil {
			//TODO ошибка в модалку
			client.Ack(*evt.Request)

			logrus.Error("Failed to collect code review context:", err)
			return
		}

		client.Ack(*evt.Request)

		if callback.Container.Type == "message" {
			_, _, err = sl.PostMessageContext(ctx, channelId,
				slack.MsgOptionReplaceOriginal(responseUrl),
				slack.MsgOptionBlocks(slackviews.GetCrPreviewEditAndConfirmBlocks(crContext, false)...))
		}

		if callback.Container.Type == "view" {
			viewId := callback.View.ID
			newViewRequest := slackviews.GetCrPreviewModal(crContext, channelId, true)

			_, err = client.UpdateViewContext(ctx, newViewRequest, "", "", viewId)
			_ = err
		}

		return
	}
}

func (h *Handler) HandleSlackViewSubmission(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback) {
	if callback.View.CallbackID != slackutils.GetCallbackId(slackviews.CrPreviewEditModal) {
		return
	}

	userId := callback.User.ID

	form := getRequestRefFromPreviewEdit(callback.View.State.Values, userId)
	request := form.ToCodeReviewCollectRequest()
	request.DisableCollectReviewersFromPr = true
	request.DisableCollectIssuesFromPr = true

	client.Ack(*evt.Request)

	err := h.handleAndSendCrThread(ctx, request, form.ChannelId, form.AsUser)
	if err != nil {
		logrus.Error("Failed to handle Slack view submission:", err)
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, form.ChannelId, callback.User.ID,
			fmt.Sprintf("Не удалось создать #CR тред: %s", err.Error()))
		return
	}
}

func (h *Handler) isCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/cr" ||
		cmd.Command == "/servit" && strings.HasPrefix(cmd.Text, "cr")
}

func getRequestRefFromPreviewEdit(state slackutils.ViewStateValues, userId string) *CrPreviewFormData {
	reviewerSlackIds := slackutils.GetSelectedUsers(state, slackviews.CrPreviewEdit, slackviews.ReviewersField)
	prUrls := slackutils.GetMultilineInputText(state, slackviews.CrPreviewEdit, slackviews.PullRequestRawListField)
	issuesUrls := slackutils.GetMultilineInputText(state, slackviews.CrPreviewEdit, slackviews.IssueRawListField)
	channelId := slackutils.GetSelectedChannel(state, slackviews.CrPreviewEdit, slackviews.ChannelField)

	return &CrPreviewFormData{
		ChannelId:        channelId,
		PullRequestUrls:  prUrls,
		IssueUrls:        issuesUrls,
		RequesterSlackId: userId,
		ReviewerSlackIds: reviewerSlackIds,
	}
}

func (h *Handler) handleAndSendCrThread(
	ctx context.Context,
	crRequest *models.CodeReviewCollectRequest,
	channelId string,
	sendAsUser bool) error {

	crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, crRequest)
	if err != nil {
		return err
	}

	messageBlocks := slackviews.GetCrThreadBlocks(crContext)

	notificationText := fmt.Sprintf("#CR от <@%s> по задаче %s", crRequest.Requester.SlackId, crContext.KeyedIssue)

	options := []slack.MsgOption{
		slack.MsgOptionText(notificationText, false),
		slack.MsgOptionBlocks(messageBlocks...),
		slack.MsgOptionMetadata(getSlackMetaData(crThreadCreatedMetaEventType, crContext)),
	}

	userProfile, err := h.bag.Client.Slack.GetUserProfileContext(ctx, &slack.GetUserProfileParameters{
		UserID: crRequest.Requester.SlackId,
	})
	if err != nil {
		logrus.Error("Failed to get Slack user info:", err)
	}

	if sendAsUser && userProfile != nil {
		options = append(options,
			slack.MsgOptionIconURL(userProfile.Image192),
			slack.MsgOptionUsername(userProfile.RealName))
	}

	respChannel, respTs, err := h.bag.Client.Slack.PostMessage(channelId, options...)
	if err != nil {
		return err
	}

	dbRecord := &models.CodeReviewThread{
		Thread: &models.ThreadRef{
			ChannelId: respChannel,
			Ts:        respTs,
		},
		Context: crContext,
		Status:  models.CodeReviewStatusOpen,
	}

	messageLink, err := h.bag.Client.Slack.GetPermalinkContext(ctx, &slack.PermalinkParameters{
		Channel: respChannel,
		Ts:      respTs,
	})
	if err == nil {
		dbRecord.MessageLink = messageLink
	}

	_, err = h.bag.DB.Repository.CreateCodeReviewThread(ctx, dbRecord)
	if err != nil {
		return err
	}

	return nil
}

func getSlackMetaData(eventType string, context *models.CodeReviewContext) slack.SlackMetadata {
	return slack.SlackMetadata{
		EventType: eventType,
		EventPayload: map[string]any{
			handlerNameMetaField: handlerNameMetaValue,
			contextMetaField:     context,
		},
	}
}
