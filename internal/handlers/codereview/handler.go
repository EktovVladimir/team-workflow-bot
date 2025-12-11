package codereview

import (
	"context"
	"fmt"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/internal/models"

	"github.com/google/go-github/v79/github"
	"github.com/samber/lo"
	"github.com/slack-go/slack"
)

// Золотистый
type retriever interface {
	CollectCodeReviewContextFromSlack(ctx context.Context, request models.RequestRef) (*models.CodeReviewContext, error)
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
	if !h.IsCommandApplicable(cmd) {
		return
	}

	// Приняли команду, команда относится к этому обработчику. Подтверждаем
	ack()

	args := strings.Fields(cmd.Text)

	if len(args) == 0 {
		//TODO модал или эфемер с инпутами
		h.sendErrorMessage(
			ctx,
			cmd.ChannelID,
			cmd.UserID,
			"Временно не поддерживается. Используй `/servit cr <pr_url1> ... <pr_urlN>`")
		return
	}

	prRefs := lo.FilterMap(args, func(arg string, i int) (*models.PullRequestRef, bool) {
		return models.ParsePullRequestRefFromUrl(arg)
	})

	if len(prRefs) == 0 {
		h.sendErrorMessage(
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

	crContext, err := h.retriever.CollectCodeReviewContextFromSlack(ctx, *request)

	if err != nil {
		h.sendErrorMessage(
			ctx,
			cmd.ChannelID,
			cmd.UserID,
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

func (h *Handler) IsCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/cr" ||
		cmd.Command == "/servit" && strings.HasPrefix(cmd.Text, "cr")
}

func (h *Handler) sendErrorMessage(
	ctx context.Context,
	channel string,
	userId string,
	text string) {
	_, _, _, _ = h.bag.Client.Slack.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionPostEphemeral(userId),
		slack.MsgOptionAttachments(
			slack.Attachment{
				Color: "danger",
				Text:  text,
			},
		),
	)
}

//TODO remove example

func (h *Handler) sendExampleMessage(ctx context.Context, cmd slack.SlashCommand) {
	slackRequester, err := h.bag.Client.Slack.GetUserInfoContext(ctx, cmd.UserID)
	if err != nil {
		return
	}

	msgBlocks := GetExample(slackRequester.Name)

	_, _, _, _ = h.bag.Client.Slack.SendMessageContext(
		ctx,
		cmd.ChannelID,
		//TODO текст уведомления
		slack.MsgOptionText("Запрос код-ревью", false),
		slack.MsgOptionIconURL(slackRequester.Profile.Image192),
		slack.MsgOptionUsername(slackRequester.RealName),
		slack.MsgOptionBlocks(msgBlocks...))
}

func GetExample(requesterSlackName string) []slack.Block {
	contextText := fmt.Sprintf("Создано через бота по запросу @%s", requesterSlackName)

	return []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "*#cr* @ivanov", false, false),
			nil,
			nil),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", ":jira: <https://aviasales.atlassian.net/browse/OTAB-4413|Удаляются записи реестра при смене типа реестра>", false, false),
			nil,
			nil),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", ":git-hub: <https://github.com/KosyanMedia/ota-flight-registry/pull/2092|OTAB-4413 += scoped transactions & use on change reg type> (FR)", false, false),
			nil,
			nil),
		slack.NewContextBlock("",
			slack.NewTextBlockObject("mrkdwn", contextText, false, false),
		),
	}
}
