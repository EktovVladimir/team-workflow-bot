package slackutils

import (
	"strings"
	"team-workflow-bot/pkg/commonutils"
	"time"

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

func GetSelectedChannel(vs ViewStateValues, base string, field string) string {
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return ""
	}

	return input.SelectedChannel
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

// GetDatePickerValue возвращает выбранную дату в datepicker в виде time.Time.
// Если значение не найдено, возвращается текущее время.
func GetDatePickerValue(vs ViewStateValues, base string, field string) time.Time {
	now := time.Now()
	input, _, _, ok := GetViewStateValue(vs, base, field)
	if !ok {
		return now
	}

	res, err := time.Parse(time.DateOnly, input.SelectedDate)
	if err != nil {
		return now
	}
	return res
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
	return initialBlockId + "_" + lo.RandomString(6, commonutils.AlphabetRunes)
}

func GetBlockIdWithRandSuffix(base string, field string) string {
	initialBlockId := GetBlockId(base, field)
	return initialBlockId + "_" + lo.RandomString(6, commonutils.AlphabetRunes)
}

func GetCallbackId(base string) string {
	return base + "_callback_id"
}
