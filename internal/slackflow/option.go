package slackflow

type ListenerOption func(cfg *listenerOptionConfig)

type listenerOptionConfig struct {
	commandHandlers []SlackSlashCommandHandler
}

func WithCommandHandler(handlers ...SlackSlashCommandHandler) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		cfg.commandHandlers = append(cfg.commandHandlers, handlers...)
	}
}

func WithAnyHandler(handlers ...any) ListenerOption {
	return func(cfg *listenerOptionConfig) {
		for _, handler := range handlers {
			if h, ok := handler.(SlackSlashCommandHandler); ok {
				cfg.commandHandlers = append(cfg.commandHandlers, h)
			}
		}
	}
}

func (cfg *listenerOptionConfig) setOptions(opts ...ListenerOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}
