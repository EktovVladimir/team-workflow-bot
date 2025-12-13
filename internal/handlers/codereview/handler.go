package codereview

import (
	"context"
	"fmt"
	"log"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/internal/models"

	"github.com/google/go-github/v79/github"
	"github.com/samber/lo"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	crThreadCreatedMetaEventType = "cr_request_created"
	handlerNameMetaField         = "handler"
	contextMetaField             = "context"
	handlerNameMetaValue         = "codereview.Handler"
)

// Золотистый
type retriever interface {
	CollectCodeReviewContextFromSlack(ctx context.Context, request *models.CodeReviewCollectRequest) (*models.CodeReviewContext, error)
}

type Handler struct {
	bag       *bag.DependenciesBag
	retriever retriever
}

func NewHandler(bag *bag.DependenciesBag, retriever retriever) *Handler {
	return &Handler{
		bag:       bag,
		retriever: retriever,
	}
}

func (h *Handler) HandlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent) {
}

func (h *Handler) HandlePullRequestReviewEvent(ctx context.Context, event *github.PullRequestReviewEvent) {
}

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
			slackviews.GetCrPreviewModal(crContext, false))
		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
				fmt.Sprintf("Ошибка при открытии модального окна превью код-ревью: %s", err.Error()))
			return
		}
	} else {
		notificationText := fmt.Sprintf("Превью #CR треда %s", crContext.Key)

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
		//TODO конфиг или доп.опция
		useRequesterIdentity = false

		crPreviewSubmitActionId    = slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.CrPreviewEditConfirm)
		crPreviewCancelActionId    = slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.CrPreviewEditCancel)
		crPreviewPrListActionId    = slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.PullRequestRawListField)
		crPreviewIssueListActionId = slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.IssueRawListField)
	)

	if actionId == crPreviewSubmitActionId {
		client.Ack(*evt.Request)

		request := getRequestRefFromPreviewEdit(callback.BlockActionState.Values, userId)
		request.DisableCollectReviewersFromPr = true
		request.DisableCollectIssuesFromPr = true

		crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)
		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, channelId, callback.User.ID,
				fmt.Sprintf("Ошибка при сборе контекста код-ревью: %s", err.Error()))
			return
		}

		messageBlocks := slackviews.GetCrThreadBlocks(crContext)

		notificationText := fmt.Sprintf("#CR от <@%s> по задаче %s", userId, crContext.Key)

		options := []slack.MsgOption{
			slack.MsgOptionText(notificationText, false),
			slack.MsgOptionBlocks(messageBlocks...),
			slack.MsgOptionMetadata(getSlackMetaData(crThreadCreatedMetaEventType, crContext)),
		}

		userProfile, err := h.bag.Client.Slack.GetUserProfileContext(ctx, &slack.GetUserProfileParameters{
			UserID: userId,
		})
		if err != nil {
			log.Println("Failed to get Slack user info:", err)
		}

		if useRequesterIdentity && userProfile != nil {
			options = append(options,
				slack.MsgOptionIconURL(userProfile.Image192),
				slack.MsgOptionUsername(userProfile.RealName))
		}

		_, _, err = sl.PostMessage(channelId, options...)
		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, channelId, callback.User.ID,
				fmt.Sprintf("Ошибка при создании треда код-ревью: %s", err.Error()))
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

		request := getRequestRefFromPreviewEdit(values, userId)
		request.DisableCollectReviewersFromPr = true

		//TODO false для crPreviewPrListActionId, но сначала разобраться с обновлением инпутов
		request.DisableCollectIssuesFromPr = actionId == crPreviewIssueListActionId

		crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)
		if err != nil {
			//TODO ошибка в модалку
			client.Ack(*evt.Request)

			log.Println("Failed to collect code review context:", err)
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
			newViewRequest := slackviews.GetCrPreviewModal(crContext, true)

			_, err = client.UpdateViewContext(ctx, newViewRequest, "", "", viewId)
			_ = err
		}

		return
	}
}

func (h *Handler) HandleSlackViewSubmission(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback) {
	if callback.View.CallbackID != slackviews.CrPreviewEditModal {
		return
	}

}

func (h *Handler) isCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/cr" ||
		cmd.Command == "/servit" && strings.HasPrefix(cmd.Text, "cr")
}

func getRequestRefFromPreviewEdit(state slackviews.ViewStateValues, userId string) *models.CodeReviewCollectRequest {
	requesterRef := &models.UserRef{
		SlackId: userId,
	}

	reviewerSlackIds := slackviews.GetSelectedUsers(state, slackviews.CrPreviewEdit, slackviews.ReviewersField)
	reviewerRefs := lo.Map(reviewerSlackIds, func(id string, i int) *models.UserRef {
		return &models.UserRef{
			SlackId: id,
		}
	})

	prUrls := slackviews.GetMultilineInputText(state, slackviews.CrPreviewEdit, slackviews.PullRequestRawListField)
	prRefs := getParsedFromUrlPullRequestRefs(prUrls)

	issuesUrls := slackviews.GetMultilineInputText(state, slackviews.CrPreviewEdit, slackviews.IssueRawListField)
	issueRefs := getParsedFromUrlIssueRefs(issuesUrls)

	return &models.CodeReviewCollectRequest{
		PullRequests: prRefs,
		Issues:       issueRefs,
		Requester:    requesterRef,
		Reviewers:    reviewerRefs,
	}
}

func getParsedFromUrlPullRequestRefs(prUrls []string) []*models.PullRequestRef {
	uniqUrls := lo.Uniq(prUrls)
	return lo.FilterMap(uniqUrls, func(arg string, i int) (*models.PullRequestRef, bool) {
		return models.ParsePullRequestRefFromUrl(arg)
	})
}

func getParsedFromUrlIssueRefs(issueUrls []string) []*models.IssueRef {
	uniqUrls := lo.Uniq(issueUrls)
	return lo.FilterMap(uniqUrls, func(arg string, i int) (*models.IssueRef, bool) {
		return models.ParseIssueRefFromUrl(arg)
	})
}

func getParsedFromFormatUserRefs(slackUserMention []string) []*models.UserRef {
	uniqMentions := lo.Uniq(slackUserMention)
	return lo.FilterMap(uniqMentions, func(arg string, i int) (*models.UserRef, bool) {
		userId, _, ok := slackflow.ParseEscapedLink(arg)
		if !ok {
			return nil, false
		}
		return &models.UserRef{
			SlackId: userId,
		}, true
	})
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
