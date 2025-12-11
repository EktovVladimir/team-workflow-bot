package slackflow

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type AckCallback func(payload ...any)

type SlackSlashCommandHandler interface {
	HandleSlackSlashCommand(ctx context.Context, cmd slack.SlashCommand, ack AckCallback)
}

type SlackDirectMessageEventHandler interface {
	HandleSlackDirectMessageEvent(ctx context.Context, event *slackevents.MessageEvent)
}

type SLackViewSubmissionHandler interface {
	HandleSlackViewSubmission(ctx context.Context, event slack.InteractionCallback, ack AckCallback)
}

type SlackBlockActionHandler interface {
	HandleSlackBlockAction(ctx context.Context, event slack.InteractionCallback, ack AckCallback)
}
