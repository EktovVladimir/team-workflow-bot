package slackflow

type BlockHelperOption func(cfg *blockHelperOptionConfig)

type blockHelperOptionConfig struct {
	hint          string
	optional      bool
	emoji         bool
	initialValue  string
	initialValues []string
	placeholder   string
}

func WithHint(hint string) BlockHelperOption {
	return func(cfg *blockHelperOptionConfig) {
		cfg.hint = hint
	}
}

func WithOptional(optional bool) BlockHelperOption {
	return func(cfg *blockHelperOptionConfig) {
		cfg.optional = optional
	}
}

func WithInitialValue(initialValue string) BlockHelperOption {
	return func(cfg *blockHelperOptionConfig) {
		cfg.initialValue = initialValue
	}
}

func WithInitialValues(initialValues []string) BlockHelperOption {
	return func(cfg *blockHelperOptionConfig) {
		cfg.initialValues = initialValues
	}
}

func WithPlaceholder(placeholder string) BlockHelperOption {
	return func(cfg *blockHelperOptionConfig) {
		cfg.placeholder = placeholder
	}
}

func WithEmoji(emoji bool) BlockHelperOption {
	return func(cfg *blockHelperOptionConfig) {
		cfg.emoji = emoji
	}
}

func applyBlockHelperOptions(opts ...BlockHelperOption) *blockHelperOptionConfig {
	cfg := &blockHelperOptionConfig{}
	for _, o := range opts {
		if o != nil {
			o(cfg)
		}
	}
	return cfg
}
