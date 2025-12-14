package slackviews

import (
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/models"
	"team-workflow-bot/pkg/slackutils"

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
		slackutils.GetUserInputBlock(
			slackutils.GetBlockId(UserEditModal, SlackField),
			slackutils.GetActionId(UserEditModal, SlackField),
			"Slack",
			slackutils.WithInitialValue(user.SlackId),
			slackutils.WithPlaceholder("Пользователь slack (обязательно)")),
		slackutils.GetTextInputBlock(
			slackutils.GetBlockId(UserEditModal, EmailField),
			slackutils.GetActionId(UserEditModal, EmailField),
			"Email",
			slackutils.WithInitialValue(user.Email),
			slackutils.WithPlaceholder("Рабочий email Котлеги (необязательно)"),
			slackutils.WithOptional(true)),
		slackutils.GetTextInputBlock(
			slackutils.GetBlockId(UserEditModal, GithubField),
			slackutils.GetActionId(UserEditModal, GithubField),
			"Github",
			slackutils.WithInitialValue(user.GitHubLogin),
			slackutils.WithPlaceholder("Логин на Github (необязательно, но очень желательно)"),
			slackutils.WithOptional(true)),
		slackutils.GetMultiSelectInputBlock(
			slackutils.GetBlockId(UserEditModal, TeamsField),
			slackutils.GetActionId(UserEditModal, TeamsField),
			"Команды",
			GetAvailableTeamsOptionValues(global.GetStorage().AvailableTeams),
			slackutils.WithInitialValues(user.Teams),
			slackutils.WithOptional(true)),
	}

	if adminMode {
		blocks = append(blocks, slackutils.GetMultiSelectInputBlock(
			slackutils.GetBlockId(UserEditModal, RolesField),
			slackutils.GetActionId(UserEditModal, RolesField),
			"Роли",
			GetAvailableRolesOptionValues(global.GetStorage().AvailableRoles),
			slackutils.WithInitialValues(user.Roles),
			slackutils.WithOptional(true)))
	}

	return slack.ModalViewRequest{
		CallbackID: slackutils.GetCallbackId(UserEditModal),
		Type:       slack.VTModal,
		Title:      slackutils.GetEmojiPlainTextObject(title),
		Submit:     slackutils.GetSimplePlainTextObject("Сохранить"),
		Close:      slackutils.GetSimplePlainTextObject("Отмена"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}
