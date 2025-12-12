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
)

const (
	crThreadCreatedMetaEventType = "cr_request_created"
	handlerNameMetaField         = "handler"
	contextMetaField             = "context"
	handlerNameMetaValue         = "codereview.Handler"
)

// Золотистый
type retriever interface {
	CollectCodeReviewContextFromSlack(ctx context.Context, request *models.RequestRef) (*models.CodeReviewContext, error)
	CollectCodeReviewContextFromSlackLite(ctx context.Context, request *models.RequestRef) (*models.CodeReviewContext, error)
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

func (h *Handler) HandleSlackSlashCommand(ctx context.Context, cmd slack.SlashCommand, ack slackflow.AckCallback) {
	if !h.isCommandApplicable(cmd) {
		return
	}

	// Приняли команду, команда относится к этому обработчику. Подтверждаем
	ack()

	args := strings.Fields(cmd.Text)

	if len(args) == 0 {
		//TODO модал или эфемер с инпутами
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
			"Временно не поддерживается. Используй `/servit cr <pr_url1> ... <pr_urlN>`")
		return
	}

	prRefs := lo.FilterMap(args, func(arg string, i int) (*models.PullRequestRef, bool) {
		return models.ParsePullRequestRefFromUrl(arg)
	})

	if len(prRefs) == 0 {
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(
			ctx,
			cmd.ChannelID,
			cmd.UserID,
			"Не удалось распознать ни одного валидного URL ПРа.")
		return
	}

	requester := &models.UserRef{
		SlackId: cmd.UserID,
	}
	request := &models.RequestRef{
		PullRequests: prRefs,
		Requester:    requester,
	}

	crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, request)

	if err != nil {
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, cmd.ChannelID, cmd.UserID,
			fmt.Sprintf("Ошибка при сборе контекста код-ревью: %s", err.Error()))
		return
	}

	//TODO пытаемся сохранить в БД юзеров

	_, _, _, err = h.bag.Client.Slack.SendMessageContext(
		ctx,
		cmd.ChannelID,
		//TODO текст уведомления
		slack.MsgOptionText("Превью запроса код-ревью", false),
		slack.MsgOptionPostEphemeral(cmd.UserID),
		slack.MsgOptionBlocks(slackviews.GetCrPreviewEditAndConfirmBlocks(crContext)...))
}

func (h *Handler) HandleSlackBlockAction(ctx context.Context, event slack.InteractionCallback, ack slackflow.AckCallback) {
	sl := h.bag.Client.Slack

	if len(event.ActionCallback.BlockActions) == 0 {
		return
	}

	actionId := event.ActionCallback.BlockActions[0].ActionID
	responseUrl := event.ResponseURL
	channelId := event.Channel.ID
	userId := event.User.ID

	if actionId == slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.CrPreviewEditConfirm) {
		ack()

		request := getRequestRefFromPreviewEdit(event.BlockActionState.Values, userId)

		crContext, err := h.retriever.CollectCodeReviewContextFromSlackLite(ctx, request)
		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, channelId, event.User.ID,
				fmt.Sprintf("Ошибка при сборе контекста код-ревью: %s", err.Error()))
			return
		}

		messageBlocks := slackviews.GetCrThreadBlocks(crContext)

		options := []slack.MsgOption{
			slack.MsgOptionText("Запрос код-ревью", false),
			slack.MsgOptionBlocks(messageBlocks...),
			slack.MsgOptionMetadata(getSlackMetaData(crThreadCreatedMetaEventType, crContext)),
		}

		userProfile, err := h.bag.Client.Slack.GetUserProfileContext(ctx, &slack.GetUserProfileParameters{
			UserID: userId,
		})
		if err != nil {
			log.Println("Failed to get Slack user info:", err)
		}

		if userProfile != nil {
			options = append(options,
				slack.MsgOptionIconURL(userProfile.Image192),
				slack.MsgOptionUsername(userProfile.RealName))
		}

		_, _, err = sl.PostMessage(channelId, options...)
		if err != nil {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, channelId, event.User.ID,
				fmt.Sprintf("Ошибка при создании треда код-ревью: %s", err.Error()))
			return
		}

		_, _, _ = sl.PostMessage(channelId, slack.MsgOptionDeleteOriginal(responseUrl))

		return
	}

	if actionId == slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.CrPreviewEditCancel) {
		_, _, _ = sl.PostMessage(channelId, slack.MsgOptionDeleteOriginal(responseUrl))

		return
	}

	if actionId == slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.IssueRawListField) ||
		actionId == slackviews.GetActionId(slackviews.CrPreviewEdit, slackviews.PullRequestRawListField) {
		ack()

		request := getRequestRefFromPreviewEdit(event.BlockActionState.Values, userId)

		crContext, err := h.retriever.CollectCodeReviewContextFromSlackLite(ctx, request)
		if err != nil {
			log.Println("Failed to collect code review context:", err)
			return
		}

		_, _, err = sl.PostMessage(channelId,
			slack.MsgOptionReplaceOriginal(responseUrl),
			slack.MsgOptionBlocks(slackviews.GetCrPreviewEditAndConfirmBlocks(crContext)...))

		return
	}
}

func (h *Handler) isCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/cr" ||
		cmd.Command == "/servit" && strings.HasPrefix(cmd.Text, "cr")
}

func getRequestRefFromPreviewEdit(state slackviews.ViewStateValues, userId string) *models.RequestRef {
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
	prRefs := lo.FilterMap(prUrls, func(arg string, i int) (*models.PullRequestRef, bool) {
		return models.ParsePullRequestRefFromUrl(arg)
	})

	issuesUrls := slackviews.GetMultilineInputText(state, slackviews.CrPreviewEdit, slackviews.IssueRawListField)
	issueRefs := lo.FilterMap(issuesUrls, func(arg string, i int) (*models.IssueRef, bool) {
		return models.ParseIssueRefFromUrl(arg)
	})

	return &models.RequestRef{
		PullRequests: prRefs,
		Issues:       issueRefs,
		Requester:    requesterRef,
		Reviewers:    reviewerRefs,
	}
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
