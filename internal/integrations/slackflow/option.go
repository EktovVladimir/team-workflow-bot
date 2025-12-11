package slackflow

type ListenerOption func(cfg *listenerOptionConfig)

type listenerOptionConfig struct {
	commandHandlers        []SlackSlashCommandHandler
	directMessageHandlers  []SlackDirectMessageEventHandler
	viewSubmissionHandlers []SLackViewSubmissionHandler
	blockActionHandlers    []SlackBlockActionHandler
}

func WithCommandHandler(handlers ...SlackSlashCommandHandler) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		cfg.commandHandlers = append(cfg.commandHandlers, handlers...)
	}
}

func WithViewSubmissionHandler(handlers ...SLackViewSubmissionHandler) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		cfg.viewSubmissionHandlers = append(cfg.viewSubmissionHandlers, handlers...)
	}
}

func WithDirectMessageHandler(handlers ...SlackDirectMessageEventHandler) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		cfg.directMessageHandlers = append(cfg.directMessageHandlers, handlers...)
	}
}

func WithBlockActionHandler(handlers ...SlackBlockActionHandler) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		cfg.blockActionHandlers = append(cfg.blockActionHandlers, handlers...)
	}
}

func WithAnyHandler(handlers ...any) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		for _, handler := range handlers {
			if h, ok := handler.(SlackSlashCommandHandler); ok {
				cfg.commandHandlers = append(cfg.commandHandlers, h)
			}
			if h, ok := handler.(SlackDirectMessageEventHandler); ok {
				cfg.directMessageHandlers = append(cfg.directMessageHandlers, h)
			}
			if h, ok := handler.(SLackViewSubmissionHandler); ok {
				cfg.viewSubmissionHandlers = append(cfg.viewSubmissionHandlers, h)
			}
			if h, ok := handler.(SlackBlockActionHandler); ok {
				cfg.blockActionHandlers = append(cfg.blockActionHandlers, h)
			}
		}
	}
}

func newListenerOptionConfig(opts ...ListenerOption) *listenerOptionConfig {
	cfg := &listenerOptionConfig{
		commandHandlers: []SlackSlashCommandHandler{},
	}
	cfg.setOptions(opts...)
	return cfg
}

func (cfg *listenerOptionConfig) setOptions(opts ...ListenerOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}
