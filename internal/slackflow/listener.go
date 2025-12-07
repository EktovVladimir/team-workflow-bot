package slackflow

import (
	"context"
	"log"
	"team-workflow-bot/internal/config"
	"team-workflow-bot/internal/environment"

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
		socketmode.OptionDebug(environment.IsDev),
	)

	optionsConfig := newListenerOptionConfig(options...)

	return &Listener{
		SlackClient:   slackClient,
		config:        config,
		socketClient:  socketClient,
		optionsConfig: optionsConfig,
	}
}

func NewListenerAndClient(config *config.Config, options ...ListenerOption) *Listener {
	botToken := config.Slack.BotToken
	appToken := config.Slack.AppToken

	slackClient := slack.New(
		botToken,
		slack.OptionDebug(environment.IsDev),
		slack.OptionAppLevelToken(appToken),
	)

	return NewListener(slackClient, config, options...)
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
		l.socketClient.Ack(*evt.Request)
		l.handleEventApi(eventsAPIEvent)
	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			return
		}
		l.socketClient.Ack(*evt.Request)
		l.handleEventCommand(ctx, cmd)
	case socketmode.EventTypeInteractive:
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			return
		}
		l.handleEventInteraction(callback)
	default:
	}
}

func (l *Listener) handleEventApi(event slackevents.EventsAPIEvent) {

	//TODO
	log.Printf("Received event: %+v\n", event)
}

func (l *Listener) handleEventCommand(ctx context.Context, cmd slack.SlashCommand) {
	log.Printf("Received command: %+v\n", cmd)

	for _, handler := range l.optionsConfig.commandHandlers {
		go func() {
			handler.HandleSlackSlashCommand(ctx, cmd)
		}()
	}
}

func (l *Listener) handleEventInteraction(callback slack.InteractionCallback) {
	//TODO
	log.Printf("Received interaction: %+v\n", callback)

	switch callback.Type {
	case slack.InteractionTypeBlockActions:
		// See https://api.slack.com/apis/connections/socket-implement#button
	case slack.InteractionTypeShortcut:
	case slack.InteractionTypeViewSubmission:
		// See https://api.slack.com/apis/connections/socket-implement#modal
	case slack.InteractionTypeDialogSubmission:
	default:
	}
}
