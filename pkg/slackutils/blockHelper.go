package slackutils

import (
	"slices"
	"team-workflow-bot/pkg/commonutils"

	"github.com/slack-go/slack"
)

type SelectBlockOption = commonutils.Triple[string, *string, *string]

func GetUserInputBlock(blockId string, actionId string, label string, options ...BlockHelperOption) *slack.InputBlock {
	cfg := applyBlockHelperOptions(options...)

	labelElement := GetSimplePlainTextObject(label)

	var hintElement *slack.TextBlockObject
	if cfg.hint != "" {
		hintElement = GetSimplePlainTextObject(cfg.hint)
	}

	selectElement := slack.NewOptionsSelectBlockElement(slack.OptTypeUser, labelElement, actionId)

	if cfg.initialValue != "" {
		selectElement.InitialUser = cfg.initialValue
	}

	return slack.NewInputBlock(
		blockId,
		labelElement,
		hintElement,
		selectElement)
}

func GetChannelInputBlock(blockId string, actionId string, label string, options ...BlockHelperOption) *slack.InputBlock {
	cfg := applyBlockHelperOptions(options...)

	labelElement := GetSimplePlainTextObject(label)

	var hintElement *slack.TextBlockObject
	if cfg.hint != "" {
		hintElement = GetSimplePlainTextObject(cfg.hint)
	}

	selectElement := slack.NewOptionsSelectBlockElement(slack.OptTypeChannels, labelElement, actionId)

	if cfg.initialValue != "" {
		selectElement.InitialChannel = cfg.initialValue
	}

	return slack.NewInputBlock(
		blockId,
		labelElement,
		hintElement,
		selectElement)
}

func GetUserMultiSelectInputBlock(
	blockId string,
	actionId string,
	label string,
	options ...BlockHelperOption) *slack.InputBlock {
	cfg := applyBlockHelperOptions(options...)

	labelElement := GetSimplePlainTextObject(label)

	var hintElement *slack.TextBlockObject
	if cfg.hint != "" {
		hintElement = GetSimplePlainTextObject(cfg.hint)
	}

	selectElement := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, labelElement, actionId)

	if len(cfg.initialValues) > 0 {
		selectElement.InitialUsers = cfg.initialValues
	}

	return slack.NewInputBlock(
		blockId,
		labelElement,
		hintElement,
		selectElement)
}
func GetTextInputBlock(blockId string, actionId string, label string, options ...BlockHelperOption) *slack.InputBlock {
	cfg := applyBlockHelperOptions(options...)

	labelElement := GetPlainTextObject(label, options...)

	var hintElement *slack.TextBlockObject
	if cfg.hint != "" {
		hintElement = GetPlainTextObject(cfg.hint, options...)
	}

	var placeholderElement *slack.TextBlockObject
	if cfg.placeholder != "" {
		placeholderElement = GetPlainTextObject(cfg.placeholder, options...)
	}

	textInputElement := slack.NewPlainTextInputBlockElement(placeholderElement, actionId)

	if cfg.initialValue != "" {
		textInputElement.InitialValue = cfg.initialValue
	}

	textInputElement.Multiline = cfg.multiline

	block := slack.NewInputBlock(
		blockId,
		labelElement,
		hintElement,
		textInputElement)

	block.Optional = cfg.optional
	block.DispatchAction = cfg.dispatchAction

	return block
}

func GetMultiSelectInputBlock(
	blockId string,
	actionId string,
	label string,
	selectOptions []SelectBlockOption,
	options ...BlockHelperOption) *slack.InputBlock {
	cfg := applyBlockHelperOptions(options...)

	labelElement := GetPlainTextObject(label, options...)

	var hintElement *slack.TextBlockObject
	if cfg.hint != "" {
		hintElement = GetPlainTextObject(cfg.hint, options...)
	}

	var placeholderElement *slack.TextBlockObject
	if cfg.placeholder != "" {
		placeholderElement = GetPlainTextObject(cfg.placeholder, options...)
	}

	blockOptionObjects := GetOptionBlockObjects(selectOptions, options...)

	multiSelectElement := slack.NewOptionsMultiSelectBlockElement(
		slack.MultiOptTypeStatic,
		placeholderElement,
		actionId,
		blockOptionObjects...,
	)

	if len(cfg.initialValues) > 0 {
		multiSelectElement.InitialOptions = filterOptionBlockObjectBySelected(blockOptionObjects, cfg.initialValues)
	}

	block := slack.NewInputBlock(
		blockId,
		labelElement,
		hintElement,
		multiSelectElement)

	block.Optional = cfg.optional

	return block
}

func GetOptionBlockObjects(selectOptions []SelectBlockOption, options ...BlockHelperOption) []*slack.OptionBlockObject {
	var selectOptionElements []*slack.OptionBlockObject
	for _, option := range selectOptions {

		text := option.First
		if option.Second != nil && *option.Second != "" {
			text = *option.Second
		}

		textElement := GetPlainTextObject(text, options...)

		var descriptionElement *slack.TextBlockObject
		if option.Second != nil {
			descriptionElement = GetPlainTextObject(*option.Third, options...)
		}

		optionBlockObject := slack.NewOptionBlockObject(
			option.First,
			textElement,
			descriptionElement,
		)

		selectOptionElements = append(selectOptionElements, optionBlockObject)
	}
	return selectOptionElements
}

func GetPlainTextObject(text string, options ...BlockHelperOption) *slack.TextBlockObject {
	cfg := applyBlockHelperOptions(options...)

	return slack.NewTextBlockObject(slack.PlainTextType, text, cfg.emoji, false)
}

func GetMarkdownTextObject(text string, options ...BlockHelperOption) *slack.TextBlockObject {
	cfg := applyBlockHelperOptions(options...)

	return slack.NewTextBlockObject(slack.MarkdownType, text, cfg.emoji, false)
}

func GetMarkdownTextSectionBlock(text string, options ...BlockHelperOption) *slack.SectionBlock {
	return slack.NewSectionBlock(GetMarkdownTextObject(text, options...), nil, nil)
}

func GetSimpleMarkdownContextBlock(text string, options ...BlockHelperOption) *slack.ContextBlock {
	cfg := applyBlockHelperOptions(options...)

	return slack.NewContextBlock("",
		slack.NewTextBlockObject(slack.MarkdownType, text, cfg.emoji, false))
}

func GetSimplePlainTextObject(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
}

func GetEmojiPlainTextObject(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.PlainTextType, text, true, false)
}

func filterOptionBlockObjectBySelected(objects []*slack.OptionBlockObject, selectValues []string) []*slack.OptionBlockObject {
	var selectOptionElements []*slack.OptionBlockObject
	for _, o := range objects {
		if slices.Contains(selectValues, o.Value) {
			selectOptionElements = append(selectOptionElements, o)
		}
	}
	return selectOptionElements
}
