package slackflow

import (
	"context"

	"github.com/slack-go/slack"
)

type SlackSlashCommandHandler interface {
	handleSlackSlashCommand(ctx context.Context, cmd slack.SlashCommand) any
}
