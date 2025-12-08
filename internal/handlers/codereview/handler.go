package codereview

import (
	"context"
	"fmt"
	"log"
	"team-workflow-bot/internal/bag"

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
	log.Println("CodeReview Handler - HandlePullRequestEvent called")
}

func (h Handler) HandlePullRequestReviewEvent(ctx context.Context, event *github.PullRequestReviewEvent) {
	log.Println("CodeReview Handler - HandlePullRequestReviewEvent called")
}

func (h Handler) HandleSlackSlashCommand(ctx context.Context, cmd slack.SlashCommand) {
	log.Println("CodeReview Handler - HandleSlackSlashCommand called")

	if cmd.Command != "/cr" {
		return
	}

	userInfo, err := h.bag.Client.Slack.GetUserInfoContext(ctx, cmd.UserID)
	if err != nil {
		return
	}

	contextText := fmt.Sprintf("Создано через бота по запросу @%s", userInfo.Name)

	blocks := []slack.Block{
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

	_, _, _, _ = h.bag.Client.Slack.SendMessageContext(
		ctx,
		cmd.ChannelID,
		//TODO
		slack.MsgOptionText("Запрос код-ревью", false),
		slack.MsgOptionIconURL(userInfo.Profile.Image192),
		slack.MsgOptionUsername(userInfo.RealName),
		slack.MsgOptionBlocks(blocks...))
}
