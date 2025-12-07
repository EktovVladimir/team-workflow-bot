package slackflow

import (
	"context"

	"github.com/slack-go/slack"
)

type SlackSlashCommandHandler interface {
	HandleSlackSlashCommand(ctx context.Context, cmd slack.SlashCommand)
}
