package modals

import (
	"team-workflow-bot/internal/global"
	slackflow2 "team-workflow-bot/internal/integrations/slackflow"
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

func GetUserEditModal(user *models.User) slack.ModalViewRequest {
	title := "Добавить"
	if user.Id != "" {
		title = "Редактировать"
	}

	return slack.ModalViewRequest{
		CallbackID: GetCallbackId(UserEditModal),
		Type:       slack.VTModal,
		Title:      slackflow2.GetEmojiPlainTextObject(title),
		Submit:     slackflow2.GetSimplePlainTextObject("Сохранить"),
		Close:      slackflow2.GetSimplePlainTextObject("Отмена"),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slackflow2.GetUserInputBlock(
					GetBlockId(UserEditModal, SlackField),
					GetActionId(UserEditModal, SlackField),
					"Slack",
					slackflow2.WithInitialValue(user.SlackId),
					slackflow2.WithPlaceholder("Пользователь slack (обязательно)")),
				slackflow2.GetTextInputBlock(
					GetBlockId(UserEditModal, EmailField),
					GetActionId(UserEditModal, EmailField),
					"Email",
					slackflow2.WithInitialValue(user.Email),
					slackflow2.WithPlaceholder("Рабочий email Котлеги (необязательно)"),
					slackflow2.WithOptional(true)),
				slackflow2.GetTextInputBlock(
					GetBlockId(UserEditModal, GithubField),
					GetActionId(UserEditModal, GithubField),
					"Github",
					slackflow2.WithInitialValue(user.GitHubLogin),
					slackflow2.WithPlaceholder("Логин на Github (необязательно, но очень желательно)"),
					slackflow2.WithOptional(true)),
				slackflow2.GetMultiSelectInputBlock(
					GetBlockId(UserEditModal, TeamsField),
					GetActionId(UserEditModal, TeamsField),
					"Команды",
					GetAvailableTeamsOptionValues(global.GetStorage().AvailableTeams),
					slackflow2.WithInitialValues(user.Teams),
					slackflow2.WithOptional(true)),
				slackflow2.GetMultiSelectInputBlock(
					GetBlockId(UserEditModal, RolesField),
					GetActionId(UserEditModal, RolesField),
					"Роли",
					GetAvailableRolesOptionValues(global.GetStorage().AvailableRoles),
					slackflow2.WithInitialValues(user.Roles),
					slackflow2.WithOptional(true)),
			},
		},
	}
}
