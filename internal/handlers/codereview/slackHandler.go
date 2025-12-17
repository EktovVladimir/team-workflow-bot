package codereview

import (
	"context"
	"fmt"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/internal/models"
	"team-workflow-bot/pkg/slackutils"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const (
	crThreadCreatedMetaEventType = "cr_request_created"
	handlerNameMetaField         = "handler"
	contextMetaField             = "context"
	handlerNameMetaValue         = "codereview.Handler"
)

var (
	crPreviewSubmitActionId    = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.CrEditConfirmAndPost)
	crPreviewCancelActionId    = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.CrEditCancel)
	crPreviewPrListActionId    = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.PullRequestRawListField)
	crPreviewIssueListActionId = slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.IssueRawListField)

	crThreadMenuShowUpdateActionId = slackutils.GetActionId(slackviews.CrThreadContextMenu, slackviews.CrShowUpdateForm)
	crThreadMenuShowDeleteActionId = slackutils.GetActionId(slackviews.CrThreadContextMenu, slackviews.CrShowDeleteConfirm)

	crThreadShowContextMenuActionId = slackutils.GetActionId(slackviews.CrThread, slackviews.CrThreadShowContextMenu)

	crPreviewModalSubmitCallbackId      = slackutils.GetCallbackId(slackviews.CrPreviewEditModal)
	crThreadMenuDeleteConfirmCallbackId = slackutils.GetCallbackId(slackviews.CrShowDeleteConfirm)
)

// TODO конфиг
const (
	sendAsModal = true
)

type SlackHandler struct {
	*Handler
}

func NewSlackHandler(bag *bag.DependenciesBag, retriever retriever) *SlackHandler {
	return &SlackHandler{
		Handler: newHandler(bag, retriever),
	}
}

func (h *SlackHandler) IsCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/cr" ||
		cmd.Command == "/servit" && strings.HasPrefix(cmd.Text, "cr")
}

func (h *SlackHandler) HandleSlackSlashCommand(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, cmd slack.SlashCommand) {
	if !h.IsCommandApplicable(cmd) {
		return
	}

	h.handleCrStartRequest(ctx, evt, client, cmd, sendAsModal)
}

func (h *SlackHandler) HandleSlackBlockAction(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback) {
	if len(callback.ActionCallback.BlockActions) == 0 {
		return
	}

	channelId := callback.Channel.ID
	userId := callback.User.ID
	threadTs := callback.Container.ThreadTs
	triggerId := callback.TriggerID

	actionId := callback.ActionCallback.BlockActions[0].ActionID

	if actionId == crPreviewSubmitActionId {
		h.handlePreviewSubmitAndPostFromEphemeral(ctx, evt, client, callback)
		return
	}

	if actionId == crPreviewCancelActionId {
		h.handleCancelFromEphemeral(ctx, evt, client, callback)
		return
	}

	if actionId == crPreviewPrListActionId || actionId == crPreviewIssueListActionId {
		h.handlePreviewChanged(ctx, evt, client, callback)
		return
	}

	if actionId == crThreadMenuShowUpdateActionId {
		//TODO
		return
	}

	if actionId == crThreadMenuShowDeleteActionId {
		h.handleThreadMenuShowDeleteConfirm(ctx, evt, client, triggerId, channelId, threadTs, userId)
	}

	if actionId == crThreadShowContextMenuActionId {
		h.handleShowThreadContextMenu(ctx, evt, client, channelId, threadTs, userId)
		return
	}
}

func (h *SlackHandler) HandleSlackViewSubmission(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback) {
	userId := callback.User.ID

	if callback.View.CallbackID == crPreviewModalSubmitCallbackId {
		h.handleSubmitAndPostFromModal(ctx, evt, client, callback)
		return
	}

	if callback.View.CallbackID == crThreadMenuDeleteConfirmCallbackId {
		threadRef := models.ParseMessageRefFromKey(callback.View.PrivateMetadata)
		h.handleDeleteCrThreadRequest(ctx, evt, client, threadRef.ChannelId, threadRef.Ts, userId)
		return
	}

}

// handleCrStartRequest Обработка команды /cr или /servit cr
// Потенциально логика может вызываться не только из команды,
// тогда потребуется заменить параметр cmd на общую модель с контекстом.
func (h *SlackHandler) handleCrStartRequest(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	cmd slack.SlashCommand,
	sendAsModal bool) {

	log := logrus.WithField("command", cmd.Command).
		WithField("user", cmd.UserID).
		WithField("channel", cmd.ChannelID)

	log.Infof("Command %s recieved with text: %s", cmd.Command, cmd.Text)

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

	triggerId := cmd.TriggerID
	openedViewId := ""

	if sendAsModal {
		viewRs, err := client.OpenViewContext(
			ctx,
			triggerId,
			slackviews.GetCrPreviewLoadingModal())
		if err != nil {
			log.Errorf("Failed to open modal view: %v", err)
			h.slackService.SendEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
				fmt.Sprintf("Ошибка при открытии модального окна превью код-ревью: %s", err.Error()))
			return
		}

		openedViewId = viewRs.View.ID
	}

	log.Debug("Collection code review context...")

	crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)

	if err != nil {
		log.Errorf("Failed to collect code review context: %v", err)
		h.slackService.SendEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
			fmt.Sprintf("Ошибка при сборе контекста код-ревью: %s", err.Error()))
		return
	}

	log.Debug("Collection code review context completed.")

	if sendAsModal {
		_, err = client.UpdateViewContext(
			ctx,
			slackviews.GetCrPreviewModal(crContext, cmd.ChannelID, false),
			"", "", openedViewId)
		if err != nil {
			log.Errorf("Failed to open modal view: %v", err)
			h.slackService.SendEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
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
			log.Errorf("Failed to send message to user: %v", err)
			h.slackService.SendEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
				fmt.Sprintf("Ошибка при отправке превью код-ревью: %s", err.Error()))
			return
		}
	}
}

// handleSubmitAndPostFromModal Подтверждение данные из превью модалки и создание треда #CR
func (h *SlackHandler) handleSubmitAndPostFromModal(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	callback slack.InteractionCallback) {

	userId := callback.User.ID

	form := h.getRequestRefFromPreviewEdit(callback.View.State.Values, slackviews.CrPreviewEdit, userId)
	request := form.ToCodeReviewCollectRequest()
	request.DisableCollectReviewersFromPr = true
	request.DisableCollectIssuesFromPr = true

	client.Ack(*evt.Request)

	err := h.postCrThread(ctx, client, request, form.ChannelId, form.AsUser)
	if err != nil {
		logrus.Error("Failed to handle Slack view submission:", err)
		h.slackService.SendEphemeralErrorMessage(ctx, form.ChannelId, callback.User.ID,
			fmt.Sprintf("Не удалось создать #CR тред: %s", err.Error()))
		return
	}
}

// handlePreviewSubmitAndPostFromEphemeral Скорее всего не будет использоваться.
// Актуальная логика в handleSubmitAndPostFromModal
func (h *SlackHandler) handlePreviewSubmitAndPostFromEphemeral(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	callback slack.InteractionCallback) {

	client.Ack(*evt.Request)

	channelId := callback.Channel.ID
	userId := callback.User.ID
	responseUrl := callback.ResponseURL

	form := h.getRequestRefFromPreviewEdit(callback.BlockActionState.Values, slackviews.CrPreviewEdit, userId)
	request := form.ToCodeReviewCollectRequest()
	request.DisableCollectReviewersFromPr = true
	request.DisableCollectIssuesFromPr = true

	err := h.postCrThread(ctx, client, request, channelId, form.AsUser)
	if err != nil {
		logrus.Error("Failed to handle and send cr thread:", err)
		h.slackService.SendEphemeralErrorMessage(ctx, channelId, callback.User.ID,
			fmt.Sprintf("Не удалось создать #CR тред: %s", err.Error()))
		return
	}

	_, _, _ = client.PostMessageContext(ctx, channelId, slack.MsgOptionDeleteOriginal(responseUrl))
}

func (h *SlackHandler) handleCancelFromEphemeral(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	callback slack.InteractionCallback) {

	channelId := callback.Channel.ID
	responseUrl := callback.ResponseURL

	client.Ack(*evt.Request)
	if responseUrl != "" {
		_, _, _ = client.PostMessageContext(ctx, channelId, slack.MsgOptionDeleteOriginal(responseUrl))
	}
}

// handlePreviewChanged Обработка изменений в превью модалке
// Например: изменение списка ПРов или задач.
// Обновляем превью модалки (или эфемерное сообщение если используется)
func (h *SlackHandler) handlePreviewChanged(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	callback slack.InteractionCallback) {

	actionId := callback.ActionCallback.BlockActions[0].ActionID
	userId := callback.User.ID
	channelId := callback.Channel.ID
	responseUrl := callback.ResponseURL

	var values map[string]map[string]slack.BlockAction

	if callback.Container.Type == "view" {
		values = callback.View.State.Values
	} else {
		values = callback.BlockActionState.Values
	}

	form := h.getRequestRefFromPreviewEdit(values, slackviews.CrPreviewEdit, userId)
	request := form.ToCodeReviewCollectRequest()

	isIssueListChanged := actionId == slackutils.GetActionId(slackviews.CrPreviewEdit, slackviews.IssueRawListField)

	request.DisableCollectReviewersFromPr = true
	request.DisableCollectIssuesFromPr = isIssueListChanged

	crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)
	if err != nil {
		//TODO ошибка в модалку
		client.Ack(*evt.Request)

		logrus.Error("Failed to collect code review context:", err)
		return
	}

	client.Ack(*evt.Request)

	if callback.Container.Type == "message" {
		_, _, err = client.PostMessageContext(ctx, channelId,
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

func (h *SlackHandler) handleShowThreadContextMenu(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	channelId string,
	ts string,
	userId string) {

	client.Ack(*evt.Request)

	//TODO проверка на пермишены

	blocks := slackviews.GetCrThreadContextMenuBlocks()

	_, err := client.PostEphemeralContext(ctx, channelId, userId,
		slack.MsgOptionText("Меню управления #CR тредом:", false),
		slack.MsgOptionTS(ts),
		slack.MsgOptionBlocks(blocks...))
	if err != nil {
		logrus.Error("Failed to send CR thread context menu:", err)
		h.slackService.SendThreadEphemeralErrorMessage(ctx, channelId, ts, userId,
			"Ошибка при открытии контекстного меню #CR треда. "+err.Error())
		return
	}
}

// handleCrUpdateRequest Обработка упоминания бота с командой /cr в треде #CR
// Позволяет отредактировать или удалить опубликованный #CR тред.
// Потенциально может вызываться не только из меншона,
// тогда потребуется заменить параметр message на общую модель с контекстом.
func (h *SlackHandler) handleCrUpdateRequest(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	message *slackevents.AppMentionEvent) {

	ts := message.ThreadTimeStamp
	channelId := message.Channel
	userId := message.User

	client.Ack(*evt.Request)

	threadRef := models.NewMessageRef(channelId, ts)

	dbCrThread, err := h.repository.GetCodeReviewThreadByThreadRef(ctx, threadRef)
	if err != nil {
		logrus.Error("Failed to get code review thread by thread ref:", err)

		h.slackService.SendEphemeralErrorMessage(ctx, channelId, userId,
			"Не удалось найти информацию В БД. "+err.Error())

		return
	}

	//TODO
	_ = dbCrThread
}

func (h *SlackHandler) handleThreadMenuShowDeleteConfirm(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	triggerId string,
	channelId string,
	ts string,
	userId string) {

	client.Ack(*evt.Request)

	_, err := h.slackService.ShowOptionsModal(ctx, triggerId, crThreadMenuDeleteConfirmCallbackId, &slackflow.ShowOptionsModalParams{
		Title: "Удалить тред #CR?",
		Text: ":warning: Будет удалена запись из БД и само сообщение в слак. Операция необратима.\n" +
			"Если в треде есть пользовательские сообщения, они удалены не будут.",
		ConfirmText: "Удалить",
		CancelText:  "Отмена",
		MetaData:    models.NewMessageRef(channelId, ts).Key,
	})
	if err != nil {
		logrus.Error("Failed to show delete CR thread confirm modal:", err)
		h.slackService.SendThreadEphemeralErrorMessage(ctx, channelId, ts, userId,
			"Не удалось открыть модальное окно "+err.Error())
		return
	}
}

func (h *SlackHandler) handleDeleteCrThreadRequest(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	channelId string,
	ts string,
	userId string) {

	client.Ack(*evt.Request)
	threadRef := models.NewMessageRef(channelId, ts)

	//TODO проверка метаданных, что это CR тред
	//TODO проверка прав на удаление

	err := h.deleteCrThread(ctx, client, threadRef)
	if err != nil {
		h.slackService.SendThreadEphemeralErrorMessage(ctx, threadRef.ChannelId, threadRef.Ts, userId,
			"Не удалось удалить #CR тред. Ошибка: : "+err.Error())
	}
}

// postCrThread публикация треда #CR
func (h *SlackHandler) postCrThread(
	ctx context.Context,
	client *socketmode.Client,
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
		slack.MsgOptionMetadata(h.getSlackMetaData(crThreadCreatedMetaEventType, crContext)),
	}

	userProfile, err := client.GetUserProfileContext(ctx, &slack.GetUserProfileParameters{
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

	respChannel, respTs, err := client.PostMessageContext(ctx, channelId, options...)
	if err != nil {
		return err
	}

	dbRecord := &models.CodeReviewThread{
		Thread:  models.NewMessageRef(respChannel, respTs),
		Context: crContext,
		Status:  models.CodeReviewStatusOpen,
	}

	// Зарезервированное первое сообщение в треде. Сюда можно будет добавить доп. информацию
	// Также тут кнопка вызова доступных действий с #CR
	_, internalTs, err := client.PostMessageContext(ctx, channelId,
		slack.MsgOptionTS(respTs),
		slack.MsgOptionBlocks(slackviews.GetCrThreadInternalInfoBlocks()...))
	if err == nil {
		dbRecord.InternalMessages = []*models.MessageRef{
			models.NewMessageRef(respChannel, internalTs),
		}
	}

	messageLink, err := client.GetPermalinkContext(ctx, &slack.PermalinkParameters{
		Channel: respChannel,
		Ts:      respTs,
	})
	if err == nil {
		dbRecord.MessageLink = messageLink
	}

	_, err = h.repository.CreateCodeReviewThread(ctx, dbRecord)
	if err != nil {
		return err
	}

	return nil
}

func (h *SlackHandler) deleteCrThread(
	ctx context.Context,
	client *socketmode.Client,
	threadRef *models.MessageRef) error {

	dbCrThread, err := h.repository.GetCodeReviewThreadByThreadRef(ctx, threadRef)
	if err != nil {
		logrus.Error("Failed to get code review thread by thread ref:", err)
		return err
	}

	err = h.repository.DeleteCodeReviewThread(ctx, dbCrThread.Id)
	if err != nil {
		logrus.Error("Failed to delete code review thread from DB:", err)
		return err
	}

	for _, ref := range dbCrThread.InternalMessages {
		_, _, err = client.DeleteMessageContext(ctx, ref.ChannelId, ref.Ts)
		if err != nil {
			logrus.Error("Failed to delete Slack thread message:", err)
			return err
		}
	}

	_, _, err = client.DeleteMessageContext(ctx, threadRef.ChannelId, threadRef.Ts)
	if err != nil {
		logrus.Error("Failed to delete Slack message:", err)
		return err
	}

	return nil
}

func (h *SlackHandler) getRequestRefFromPreviewEdit(state slackutils.ViewStateValues, base string, userId string) *CrPreviewFormData {
	reviewerSlackIds := slackutils.GetSelectedUsers(state, base, slackviews.ReviewersField)
	prUrls := slackutils.GetMultilineInputText(state, base, slackviews.PullRequestRawListField)
	issuesUrls := slackutils.GetMultilineInputText(state, base, slackviews.IssueRawListField)
	channelId := slackutils.GetSelectedChannel(state, base, slackviews.ChannelField)

	return &CrPreviewFormData{
		ChannelId:        channelId,
		PullRequestUrls:  prUrls,
		IssueUrls:        issuesUrls,
		RequesterSlackId: userId,
		ReviewerSlackIds: reviewerSlackIds,
	}
}

func (h *SlackHandler) getSlackMetaData(eventType string, context *models.CodeReviewContext) slack.SlackMetadata {
	return slack.SlackMetadata{
		EventType: eventType,
		EventPayload: map[string]any{
			handlerNameMetaField: handlerNameMetaValue,
			contextMetaField:     context,
		},
	}
}

func (h *SlackHandler) isCodeReviewContextFull(cr *models.CodeReviewContext) bool {
	hasReviewer := lo.SomeBy(cr.Reviewers, func(reviewer *models.UserRef) bool {
		return reviewer.SlackId != ""
	})

	hasIssue := len(cr.Issues) > 0
	hasPr := len(cr.PullRequests) > 0

	return hasReviewer &&
		hasIssue &&
		hasPr &&
		cr.Requester.SlackId != ""
}
