package slackflow

import (
	"context"
	"team-workflow-bot/internal/config"
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/logger"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Listener struct {
	config       *config.Config
	socketClient *socketmode.Client
	cfg          *listenerOptionConfig

	SlackClient *slack.Client
}

func NewListener(slackClient *slack.Client, config *config.Config, options ...ListenerOption) *Listener {
	socketClient := socketmode.New(
		slackClient,
		socketmode.OptionDebug(global.IsDev),
		socketmode.OptionLog(logger.NewStdLogger(logrus.DebugLevel, "slack-socket: ")),
	)

	optionsConfig := newListenerOptionConfig(options...)

	return &Listener{
		SlackClient:  slackClient,
		config:       config,
		socketClient: socketClient,
		cfg:          optionsConfig,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	socketModeHandler := socketmode.NewSocketmodeHandler(l.socketClient)

	for _, handler := range l.cfg.commandHandlers {
		socketModeHandler.Handle(socketmode.EventTypeSlashCommand, l.slashCommandMiddleware(ctx, handler))
	}

	for _, handler := range l.cfg.blockActionHandlers {
		socketModeHandler.HandleInteraction(slack.InteractionTypeBlockActions, l.blockActionEventMiddleware(ctx, handler))
	}

	for _, handler := range l.cfg.viewSubmissionHandlers {
		socketModeHandler.HandleInteraction(slack.InteractionTypeViewSubmission, l.viewSubmissionEventMiddleware(ctx, handler))
	}

	socketModeHandler.Handle(socketmode.EventTypeEventsAPI, l.eventApiMiddleware(ctx))

	return socketModeHandler.RunEventLoopContext(ctx)
}

func (l *Listener) slashCommandMiddleware(ctx context.Context, handler SlackSlashCommandHandler) socketmode.SocketmodeHandlerFunc {
	return func(evt *socketmode.Event, client *socketmode.Client) {
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			return
		}
		handler.HandleSlackSlashCommand(ctx, evt, client, cmd)
	}
}

func (l *Listener) blockActionEventMiddleware(ctx context.Context, handler SlackBlockActionHandler) socketmode.SocketmodeHandlerFunc {
	return func(evt *socketmode.Event, client *socketmode.Client) {
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			return
		}
		handler.HandleSlackBlockAction(ctx, evt, client, callback)
	}
}

func (l *Listener) viewSubmissionEventMiddleware(ctx context.Context, handler SLackViewSubmissionHandler) socketmode.SocketmodeHandlerFunc {
	return func(evt *socketmode.Event, client *socketmode.Client) {
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			return
		}
		handler.HandleSlackViewSubmission(ctx, evt, client, callback)
	}
}

func (l *Listener) eventApiMiddleware(ctx context.Context) socketmode.SocketmodeHandlerFunc {
	return func(evt *socketmode.Event, client *socketmode.Client) {
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}

		// Можем ответить сразу, так как это не интерактивный ивент
		client.Ack(*evt.Request)

		switch eventsAPIEvent.Type {
		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				for _, handler := range l.cfg.appMentionHandlers {
					go handler.HandleSlackAppMentionEvent(ctx, evt, client, ev)
				}
			case *slackevents.MessageEvent:
				if ev.ChannelType == "im" {
					for _, handler := range l.cfg.directMessageHandlers {
						go handler.HandleSlackDirectMessageEvent(ctx, evt, client, ev)
					}
				}
			}
		default:
			client.Debugf("unsupported Events API event received")
		}
	}
}
