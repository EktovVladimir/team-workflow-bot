package githubflow

type HandlerOption func(cfg *handlerOptionConfig)

type handlerOptionConfig struct {
	prHandlers       []GithubPullRequestEventHandler
	prReviewHandlers []GithubPullRequestReviewEventHandler
	wfRunHandler     []GithubWorkflowRunEventHandler
}

func WithPullRequestHandler(handlers ...GithubPullRequestEventHandler) HandlerOption {
	return func(cfg *handlerOptionConfig) {
		cfg.prHandlers = append(cfg.prHandlers, handlers...)
	}
}

func WithPullRequestReviewHandler(handlers ...GithubPullRequestReviewEventHandler) HandlerOption {
	return func(cfg *handlerOptionConfig) {
		cfg.prReviewHandlers = append(cfg.prReviewHandlers, handlers...)
	}
}

func WithWorkflowRunHandler(handlers ...GithubWorkflowRunEventHandler) HandlerOption {
	return func(cfg *handlerOptionConfig) {
		cfg.wfRunHandler = append(cfg.wfRunHandler, handlers...)
	}
}

func WithAnyHandler(handlers ...any) HandlerOption {
	return func(cfg *handlerOptionConfig) {
		for _, handler := range handlers {
			if h, ok := handler.(GithubPullRequestEventHandler); ok {
				cfg.prHandlers = append(cfg.prHandlers, h)
			}

			if h, ok := handler.(GithubPullRequestReviewEventHandler); ok {
				cfg.prReviewHandlers = append(cfg.prReviewHandlers, h)
			}

			if h, ok := handler.(GithubWorkflowRunEventHandler); ok {
				cfg.wfRunHandler = append(cfg.wfRunHandler, h)
			}
		}
	}
}

func newHandlerOptionConfig(opts ...HandlerOption) *handlerOptionConfig {
	cfg := &handlerOptionConfig{
		prHandlers:       []GithubPullRequestEventHandler{},
		prReviewHandlers: []GithubPullRequestReviewEventHandler{},
		wfRunHandler:     []GithubWorkflowRunEventHandler{},
	}
	cfg.setOptions(opts...)
	return cfg
}

func (cfg *handlerOptionConfig) setOptions(opts ...HandlerOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}
