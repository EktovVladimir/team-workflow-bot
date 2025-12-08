package codereview

import (
	"context"
	"fmt"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/integrations/slackflow"

	"github.com/google/go-github/v79/github"
	"github.com/slack-go/slack"
)

type Handler struct {
	bag *bag.DependenciesBag
}

func NewHandler(bag *bag.DependenciesBag) *Handler {
	return &Handler{
		bag: bag,
	}
}

func (h Handler) HandlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent) {
}

func (h Handler) HandlePullRequestReviewEvent(ctx context.Context, event *github.PullRequestReviewEvent) {
}

func (h Handler) HandleSlackSlashCommand(ctx context.Context, cmd slack.SlashCommand, ack slackflow.AckCallback) {
	if !h.IsCommandApplicable(cmd) {
		return
	}

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

func (h Handler) IsCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/cr"
}

//TODO remove example

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
