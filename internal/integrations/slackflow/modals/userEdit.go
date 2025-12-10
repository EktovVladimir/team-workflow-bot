package modals

import (
	"slices"
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/integrations/slackflow"
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

	blocks := []slack.Block{
		slackflow.GetUserInputBlock(
			GetBlockId(UserEditModal, SlackField),
			GetActionId(UserEditModal, SlackField),
			"Slack",
			slackflow.WithInitialValue(user.SlackId),
			slackflow.WithPlaceholder("Пользователь slack (обязательно)")),
		slackflow.GetTextInputBlock(
			GetBlockId(UserEditModal, EmailField),
			GetActionId(UserEditModal, EmailField),
			"Email",
			slackflow.WithInitialValue(user.Email),
			slackflow.WithPlaceholder("Рабочий email Котлеги (необязательно)"),
			slackflow.WithOptional(true)),
		slackflow.GetTextInputBlock(
			GetBlockId(UserEditModal, GithubField),
			GetActionId(UserEditModal, GithubField),
			"Github",
			slackflow.WithInitialValue(user.GitHubLogin),
			slackflow.WithPlaceholder("Логин на Github (необязательно, но очень желательно)"),
			slackflow.WithOptional(true)),
		slackflow.GetMultiSelectInputBlock(
			GetBlockId(UserEditModal, TeamsField),
			GetActionId(UserEditModal, TeamsField),
			"Команды",
			GetAvailableTeamsOptionValues(global.GetStorage().AvailableTeams),
			slackflow.WithInitialValues(user.Teams),
			slackflow.WithOptional(true)),
	}

	if slices.Contains(user.Roles, "admin") {
		blocks = append(blocks, slackflow.GetMultiSelectInputBlock(
			GetBlockId(UserEditModal, RolesField),
			GetActionId(UserEditModal, RolesField),
			"Роли",
			GetAvailableRolesOptionValues(global.GetStorage().AvailableRoles),
			slackflow.WithInitialValues(user.Roles),
			slackflow.WithOptional(true)))
	}

	return slack.ModalViewRequest{
		CallbackID: GetCallbackId(UserEditModal),
		Type:       slack.VTModal,
		Title:      slackflow.GetEmojiPlainTextObject(title),
		Submit:     slackflow.GetSimplePlainTextObject("Сохранить"),
		Close:      slackflow.GetSimplePlainTextObject("Отмена"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}
