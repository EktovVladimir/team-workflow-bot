package slackviews

import (
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/models"

	"github.com/slack-go/slack"
)

const (
	UserEditModal = "user_edit_modal"
	SlackField    = "slack"
	EmailField    = "email"
	GithubField   = "github"
	RolesField    = "roles"
	TeamsField    = "teams"
)

func GetUserEditModal(user *models.User, adminMode bool) slack.ModalViewRequest {
	title := "Добавить"
	if user.Id != "" {
		title = "Редактировать"
	}

	blocks := []slack.Block{
		GetUserInputBlock(
			GetBlockId(UserEditModal, SlackField),
			GetActionId(UserEditModal, SlackField),
			"Slack",
			WithInitialValue(user.SlackId),
			WithPlaceholder("Пользователь slack (обязательно)")),
		GetTextInputBlock(
			GetBlockId(UserEditModal, EmailField),
			GetActionId(UserEditModal, EmailField),
			"Email",
			WithInitialValue(user.Email),
			WithPlaceholder("Рабочий email Котлеги (необязательно)"),
			WithOptional(true)),
		GetTextInputBlock(
			GetBlockId(UserEditModal, GithubField),
			GetActionId(UserEditModal, GithubField),
			"Github",
			WithInitialValue(user.GitHubLogin),
			WithPlaceholder("Логин на Github (необязательно, но очень желательно)"),
			WithOptional(true)),
		GetMultiSelectInputBlock(
			GetBlockId(UserEditModal, TeamsField),
			GetActionId(UserEditModal, TeamsField),
			"Команды",
			GetAvailableTeamsOptionValues(global.GetStorage().AvailableTeams),
			WithInitialValues(user.Teams),
			WithOptional(true)),
	}

	if adminMode {
		blocks = append(blocks, GetMultiSelectInputBlock(
			GetBlockId(UserEditModal, RolesField),
			GetActionId(UserEditModal, RolesField),
			"Роли",
			GetAvailableRolesOptionValues(global.GetStorage().AvailableRoles),
			WithInitialValues(user.Roles),
			WithOptional(true)))
	}

	return slack.ModalViewRequest{
		CallbackID: GetCallbackId(UserEditModal),
		Type:       slack.VTModal,
		Title:      GetEmojiPlainTextObject(title),
		Submit:     GetSimplePlainTextObject("Сохранить"),
		Close:      GetSimplePlainTextObject("Отмена"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}
