package configurator

import (
	"context"
	"errors"
	"log"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/db"
	slackflow2 "team-workflow-bot/internal/integrations/slackflow"
	modals2 "team-workflow-bot/internal/integrations/slackflow/modals"
	"team-workflow-bot/internal/models"

	"github.com/slack-go/slack"
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
	cmd slack.SlashCommand,
	ack slackflow2.AckCallback) {
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
	ack slackflow2.AckCallback) {

	if event.View.CallbackID == modals2.GetCallbackId(modals2.UserEditModal) {

		slackId := modals2.GetSelectedUser(event.View.State, modals2.UserEditModal, modals2.SlackField)
		email := modals2.GetInputText(event.View.State, modals2.UserEditModal, modals2.EmailField)
		githubName := modals2.GetInputText(event.View.State, modals2.UserEditModal, modals2.GithubField)
		roles := modals2.GetMultiSelectValues(event.View.State, modals2.UserEditModal, modals2.RolesField)
		teams := modals2.GetMultiSelectValues(event.View.State, modals2.UserEditModal, modals2.TeamsField)

		//TODO валидация

		ack()

		user, err := s.bag.DB.Repository.GetBySlackId(ctx, slackId)
		if err != nil && !errors.Is(err, db.RecordNotFound) {
			log.Printf("Error getting user by slack id %v: %v", slackId, err)
			return
		}

		if user == nil {
			user = &models.User{
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
		s.sendUserEditModal(ctx, data, &models.User{})
		return
	}

	if len(args) == 1 {
		userId, _ := slackflow2.ParseEscapedLink(args[0])

		user, err := s.bag.DB.Repository.GetBySlackId(ctx, userId)
		if err != nil && !errors.Is(err, db.RecordNotFound) {
			s.sendErrorMessage(ctx, data, "Ошибка при запросе БД: "+err.Error())
			return
		}

		if user == nil {
			user = &models.User{
				SlackId: userId,
			}
		}

		s.sendUserEditModal(ctx, data, user)
		return
	}

	//TODO парсим аргументы из текста и создаем пользователя
}

func (s SlackBotConfigurationHandler) sendUserEditModal(ctx context.Context, data interactivityData, user *models.User) {

	modals2.GetUserEditModal(user)
	_, err := s.bag.Client.Slack.OpenViewContext(ctx, data.triggerId, modals2.GetUserEditModal(user))

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
