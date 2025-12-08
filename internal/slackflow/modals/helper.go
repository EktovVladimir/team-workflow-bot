package modals

import (
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/slackflow"

	"github.com/slack-go/slack"
)

func GetSelectedUser(vs *slack.ViewState, base string, field string) string {
	return GetViewStateValue(vs, base, field).SelectedUser
}

func GetInputText(vs *slack.ViewState, base string, field string) string {
	return GetViewStateValue(vs, base, field).Value
}

func GetMultiSelectValues(vs *slack.ViewState, base string, field string) []string {
	objects := GetViewStateValue(vs, base, field).SelectedOptions
	return GetStringValuesFromObjects(objects...)
}

func GetAvailableRolesOptionValues(roles []db.Role) []slackflow.SelectBlockOption {
	var options []slackflow.SelectBlockOption

	for _, role := range roles {
		options = append(options, slackflow.SelectBlockOption{
			First:  role.Name,
			Second: &role.Name,
			Third:  &role.Description,
		})
	}

	return options
}

func GetAvailableTeamsOptionValues(teams []db.Team) []slackflow.SelectBlockOption {
	var options []slackflow.SelectBlockOption

	for _, team := range teams {
		options = append(options, slackflow.SelectBlockOption{
			First:  team.Name,
			Second: &team.Name,
			Third:  &team.Description,
		})
	}

	return options
}

func GetViewStateValue(vs *slack.ViewState, base string, field string) slack.BlockAction {
	return vs.Values[GetBlockId(base, field)][GetActionId(base, field)]
}

func GetStringValuesFromObjects(objects ...slack.OptionBlockObject) []string {
	var values []string
	for _, obj := range objects {
		values = append(values, obj.Value)
	}
	return values
}

func GetActionId(base string, field string) string {
	return base + "_" + field + "_action_id"
}

func GetBlockId(base string, field string) string {
	return base + "_" + field + "_block_id"
}

func GetCallbackId(base string) string {
	return base + "_callback_id"
}
