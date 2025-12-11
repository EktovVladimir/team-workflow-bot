package slackflow

import (
	"context"
	"log"
	"os"
	"team-workflow-bot/internal/config"
	"team-workflow-bot/internal/global"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Listener struct {
	config        *config.Config
	socketClient  *socketmode.Client
	optionsConfig *listenerOptionConfig

	SlackClient *slack.Client
}

func NewListener(slackClient *slack.Client, config *config.Config, options ...ListenerOption) *Listener {
	socketClient := socketmode.New(
		slackClient,
		socketmode.OptionDebug(global.IsDev),
		socketmode.OptionLog(log.New(os.Stdout, "slack-socket: ", log.Lshortfile|log.LstdFlags)),
	)

	optionsConfig := newListenerOptionConfig(options...)

	return &Listener{
		SlackClient:   slackClient,
		config:        config,
		socketClient:  socketClient,
		optionsConfig: optionsConfig,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	go func() {
		for evt := range l.socketClient.Events {
			l.handleEvent(ctx, evt)
		}
	}()

	err := l.socketClient.RunContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (l *Listener) handleEvent(ctx context.Context, evt socketmode.Event) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}
		// Тут можем сразу ack-нуть, так как не требуется валидация.
		l.socketClient.Ack(*evt.Request)
		l.handleEventApi(ctx, eventsAPIEvent)
	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			return
		}
		l.handleEventCommand(ctx, cmd, evt)
	case socketmode.EventTypeInteractive:
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			return
		}
		l.handleEventInteraction(ctx, callback, evt)
	default:
	}
}

func (l *Listener) handleEventApi(ctx context.Context, event slackevents.EventsAPIEvent) {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			if ev.ChannelType == "im" {
				for _, handler := range l.optionsConfig.directMessageHandlers {
					go func() {
						handler.HandleSlackDirectMessageEvent(ctx, ev)
					}()
				}
			}
		default:
		}
	}
}

func (l *Listener) handleEventCommand(ctx context.Context, cmd slack.SlashCommand, evt socketmode.Event) {
	for _, handler := range l.optionsConfig.commandHandlers {
		go func() {
			handler.HandleSlackSlashCommand(ctx, cmd, func(payload ...any) {
				l.socketClient.Ack(*evt.Request, payload...)
			})
		}()
	}
}

func (l *Listener) handleEventInteraction(ctx context.Context, callback slack.InteractionCallback, evt socketmode.Event) {
	ackCallback := func(payload ...any) {
		l.socketClient.Ack(*evt.Request, payload...)
	}

	switch callback.Type {
	case slack.InteractionTypeViewSubmission:
		for _, handler := range l.optionsConfig.viewSubmissionHandlers {
			go func() {
				handler.HandleSlackViewSubmission(ctx, callback, ackCallback)
			}()
		}
	case slack.InteractionTypeBlockActions:
		for _, handler := range l.optionsConfig.blockActionHandlers {
			go func() {
				handler.HandleSlackBlockAction(ctx, callback, ackCallback)
			}()
		}
	default:
	}
}
