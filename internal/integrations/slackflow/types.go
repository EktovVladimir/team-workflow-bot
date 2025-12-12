package slackflow

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type AckCallback func(payload ...any)

type SlackSlashCommandHandler interface {
	HandleSlackSlashCommand(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, cmd slack.SlashCommand)
}

type SlackDirectMessageEventHandler interface {
	HandleSlackDirectMessageEvent(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, message *slackevents.MessageEvent)
}

type SLackViewSubmissionHandler interface {
	HandleSlackViewSubmission(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback)
}

type SlackBlockActionHandler interface {
	HandleSlackBlockAction(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback)
}
