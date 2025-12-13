package slackviews

import (
	"strings"
	"team-workflow-bot/internal/constants"
	"team-workflow-bot/internal/models"

	"github.com/samber/lo"
	"github.com/slack-go/slack"
)

type ViewStateValues = map[string]map[string]slack.BlockAction

func GetSelectedUser(vs ViewStateValues, base string, field string) string {
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return ""
	}

	return input.SelectedUser
}

func GetSelectedUsers(vs ViewStateValues, base string, field string) []string {
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return []string{}
	}

	return input.SelectedUsers
}

func GetInputText(vs ViewStateValues, base string, field string) string {
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return ""
	}

	return input.Value
}

func GetMultilineInputText(vs ViewStateValues, base string, field string) []string {
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return []string{}
	}

	return strings.Split(input.Value, "\n")
}

func GetMultiSelectValues(vs ViewStateValues, base string, field string) []string {
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return []string{}
	}

	objects := input.SelectedOptions
	return GetStringValuesFromObjects(objects...)
}

func GetAvailableRolesOptionValues(roles []models.Role) []SelectBlockOption {
	var options []SelectBlockOption

	for _, role := range roles {
		options = append(options, SelectBlockOption{
			First:  role.Name,
			Second: &role.Name,
			Third:  &role.Description,
		})
	}

	return options
}

func GetAvailableTeamsOptionValues(teams []models.Team) []SelectBlockOption {
	var options []SelectBlockOption

	for _, team := range teams {
		options = append(options, SelectBlockOption{
			First:  team.Name,
			Second: &team.Name,
			Third:  &team.Description,
		})
	}

	return options
}

func GetViewStateValue(vs ViewStateValues, base string, field string) (slack.BlockAction, string, string, bool) {
	blockId := GetBlockId(base, field)
	actionId := GetActionId(base, field)

	block, ok := vs[blockId]

	// Можем в конце blockId добавлять суффикс, для сброса состояния (пользовательского ввода),
	// поэтому ищем блок по префиксу
	if !ok {
		actualBlockKey, ok := lo.FindKeyBy(vs, func(key string, _ map[string]slack.BlockAction) bool {
			return strings.HasPrefix(key, blockId+"_")
		})
		if ok {
			blockId = actualBlockKey
			block = vs[actualBlockKey]
		}
	}

	if block == nil {
		return slack.BlockAction{}, blockId, actionId, false
	}

	action, ok := block[actionId]

	if !ok {
		actualActionKey, ok := lo.FindKeyBy(block, func(key string, _ slack.BlockAction) bool {
			return strings.HasPrefix(key, actionId+"_")
		})
		if ok {
			actionId = actualActionKey
			action = block[actualActionKey]
		}
	}

	return action, blockId, actionId, true
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

func GetActionIdWithRandSuffix(base string, field string) string {
	initialBlockId := GetActionId(base, field)
	return initialBlockId + "_" + lo.RandomString(6, constants.AlphabetRunes)
}

func GetBlockIdWithRandSuffix(base string, field string) string {
	initialBlockId := GetBlockId(base, field)
	return initialBlockId + "_" + lo.RandomString(6, constants.AlphabetRunes)
}

func GetCallbackId(base string) string {
	return base + "_callback_id"
}
