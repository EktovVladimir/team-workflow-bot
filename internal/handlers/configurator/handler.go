package configurator

import (
	"context"
	"errors"
	"log"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/slackflow"

	"github.com/slack-go/slack"
)

const (
	userEditModal = "user_edit_modal"
	slackField    = "slack"
	emailField    = "email"
	githubField   = "github"
	rolesField    = "roles"
	teamsField    = "teams"
)

type interactivityData struct {
	channelId string
	userId    string
	triggerId string
}

type SlackBotConfigurationHandler struct {
	bag *bag.DependenciesBag
}

func NewSlackBotConfigurationHandler(
	bag *bag.DependenciesBag) *SlackBotConfigurationHandler {
	return &SlackBotConfigurationHandler{
		bag: bag,
	}
}

func (s SlackBotConfigurationHandler) HandleSlackSlashCommand(
	ctx context.Context,
	cmd slack.SlashCommand) {
	if cmd.Command != "/wfbot" {
		return
	}

	args := strings.Split(cmd.Text, " ")

	if len(args) == 0 {
		return
	}

	//TODO проверка на роль

	data := interactivityData{
		channelId: cmd.ChannelID,
		userId:    cmd.UserID,
		triggerId: cmd.TriggerID,
	}

	if len(args) >= 1 && args[0] == "user" {
		s.configureUsers(ctx, data, args[1:])
		return
	}
}

func (s SlackBotConfigurationHandler) HandleSlackViewSubmission(
	ctx context.Context,
	event slack.InteractionCallback,
	ack slackflow.AckCallback) {

	if event.View.CallbackID == getCallbackId(userEditModal) {
		slackId := getViewStateValue(event.View.State, userEditModal, slackField).SelectedUser
		email := getViewStateValue(event.View.State, userEditModal, emailField).Value
		githubName := getViewStateValue(event.View.State, userEditModal, githubField).Value
		rolesObjects := getViewStateValue(event.View.State, userEditModal, rolesField).SelectedOptions
		roles := getStringValuesFromObjects(rolesObjects...)
		teamsObjects := getViewStateValue(event.View.State, userEditModal, teamsField).SelectedOptions
		teams := getStringValuesFromObjects(teamsObjects...)

		//TODO валидация

		ack()

		user, err := s.bag.DB.Repository.GetBySlackId(ctx, slackId)
		if err != nil && !errors.Is(err, db.RecordNotFound) {
			log.Printf("Error getting user by slack id %v: %v", slackId, err)
			return
		}

		if user == nil {
			user = &db.User{
				SlackId:     slackId,
				Email:       email,
				GitHubLogin: githubName,
				Roles:       roles,
				Teams:       teams,
			}
			_, err = s.bag.DB.Repository.CreateUser(ctx, user)
			if err != nil {
				log.Printf("Error creating user with slack id %v: %v", slackId, err)
				return
			}
		} else {
			user.Email = email
			user.GitHubLogin = githubName
			user.Roles = roles
			user.Teams = teams

			err = s.bag.DB.Repository.UpdateUser(ctx, user)
			if err != nil {
				log.Printf("Error updating user with slack id %v: %v", slackId, err)
				return
			}
		}
	}
}

func (s SlackBotConfigurationHandler) configureUsers(
	ctx context.Context,
	data interactivityData,
	args []string) {

	if len(args) == 0 {
		s.sendUserEditModal(ctx, data, &db.User{})
		return
	}

	if len(args) == 1 {
		userId, _ := slackflow.ParseEscapedLink(args[0])

		user, err := s.bag.DB.Repository.GetBySlackId(ctx, userId)
		if err != nil && !errors.Is(err, db.RecordNotFound) {
			s.sendErrorMessage(ctx, data, "Ошибка при запросе БД: "+err.Error())
			return
		}

		if user == nil {
			user = &db.User{
				SlackId: userId,
			}
		}

		s.sendUserEditModal(ctx, data, user)
		return
	}

	//TODO парсим аргументы из текста и создаем пользователя
}

func (s SlackBotConfigurationHandler) sendUserEditModal(ctx context.Context, data interactivityData, user *db.User) {
	title := "Добавить Котлегу"
	if user.Id != "" {
		title = "Редактировать Котлегу"
	}

	_, err := s.bag.Client.Slack.OpenViewContext(ctx, data.triggerId, slack.ModalViewRequest{
		CallbackID: getCallbackId(userEditModal),
		Type:       slack.VTModal,
		Title:      slackflow.GetEmojiPlainTextObject(title),
		Submit:     slackflow.GetSimplePlainTextObject("Сохранить"),
		Close:      slackflow.GetSimplePlainTextObject("Отмена"),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slackflow.GetUserInputBlock(
					getBlockId(userEditModal, slackField),
					getActionId(userEditModal, slackField),
					"Slack",
					slackflow.WithInitialValue(user.SlackId),
					slackflow.WithPlaceholder("Пользователь slack (обязательно)")),
				slackflow.GetTextInputBlock(
					getBlockId(userEditModal, emailField),
					getActionId(userEditModal, emailField),
					"Email",
					slackflow.WithInitialValue(user.Email),
					slackflow.WithPlaceholder("Рабочий email Котлеги (необязательно)"),
					slackflow.WithOptional(true)),
				slackflow.GetTextInputBlock(
					getBlockId(userEditModal, githubField),
					getActionId(userEditModal, githubField),
					"Github",
					slackflow.WithInitialValue(user.GitHubLogin),
					slackflow.WithPlaceholder("Логин на Github (необязательно, но очень желательно)"),
					slackflow.WithOptional(true)),
				slackflow.GetMultiSelectInputBlock(
					getBlockId(userEditModal, teamsField),
					getActionId(userEditModal, teamsField),
					"Команды",
					getAvailableTeamsOptionValues(global.GetStorage().AvailableTeams),
					slackflow.WithInitialValues(user.Teams),
					slackflow.WithOptional(true)),
				slackflow.GetMultiSelectInputBlock(
					getBlockId(userEditModal, rolesField),
					getActionId(userEditModal, rolesField),
					"Роли",
					getAvailableRolesOptionValues(global.GetStorage().AvailableRoles),
					slackflow.WithInitialValues(user.Roles),
					slackflow.WithOptional(true)),
			},
		},
	})

	if err != nil {
		s.sendErrorMessage(ctx, data, "Не удалось открыть модальное окно для редактирования пользователя. "+err.Error())
		return
	}
}

// TODO вынести в хелпер
func (s SlackBotConfigurationHandler) sendErrorMessage(
	ctx context.Context,
	data interactivityData,
	text string) {
	_, _, _, _ = s.bag.Client.Slack.SendMessageContext(
		ctx,
		data.channelId,
		slack.MsgOptionPostEphemeral(data.userId),
		slack.MsgOptionAttachments(
			slack.Attachment{
				Color: "danger",
				Text:  text,
			},
		),
	)
}

func getAvailableRolesOptionValues(roles []db.Role) []slackflow.SelectBlockOption {
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

func getAvailableTeamsOptionValues(teams []db.Team) []slackflow.SelectBlockOption {
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

func getViewStateValue(vs *slack.ViewState, base string, field string) slack.BlockAction {
	return vs.Values[getBlockId(base, field)][getActionId(base, field)]
}

func getStringValuesFromObjects(objects ...slack.OptionBlockObject) []string {
	var values []string
	for _, obj := range objects {
		values = append(values, obj.Value)
	}
	return values
}

func getActionId(base string, field string) string {
	return base + "_" + field + "_action_id"
}

func getBlockId(base string, field string) string {
	return base + "_" + field + "_block_id"
}

func getCallbackId(base string) string {
	return base + "_callback_id"
}
