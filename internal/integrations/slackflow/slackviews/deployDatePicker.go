package slackviews

import (
	"team-workflow-bot/pkg/commonutils"
	"team-workflow-bot/pkg/slackutils"
	"time"

	"github.com/slack-go/slack"
)

const (
	DeployDatePicker = "deploy_date_picker"
	DeployDateField  = "deploy_date"
	SubmitDeployDate = "submit_deploy_date"
)

func GetDeployDatePickerBlocks(initialDate time.Time) []slack.Block {

	dataPickerElement := slack.NewDatePickerBlockElement(slackutils.GetActionId(DeployDatePicker, DeployDateField))

	dataPickerElement.InitialDate = initialDate.Format(time.DateOnly)

	actionBlock := slack.NewActionBlock(
		slackutils.GetBlockId(DeployDatePicker, DeployDateField),
		dataPickerElement,
		slack.NewButtonBlockElement(
			slackutils.GetActionId(DeployDatePicker, SubmitDeployDate),
			"submit",
			slackutils.GetPlainTextObject("Подтвердить")))

	return []slack.Block{
		slackutils.GetMarkdownTextSectionBlock("Выбери дату деплоя:"),
		actionBlock,
	}
}

func GetDeployDatePostBlocks(deplyDate time.Time) []slack.Block {
	return []slack.Block{
		slackutils.GetMarkdownTextSectionBlock("• Дата деплоя: " + deplyDate.Format(commonutils.RuDateFormat)),
	}
}
