package configurator

import (
	"context"
	"errors"
	"log"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/integrations/slackflow/salckviews"
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
	ack slackflow.AckCallback) {
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

	if event.View.CallbackID == salckviews.GetCallbackId(salckviews.UserEditModal) {

		slackId := salckviews.GetSelectedUser(event.View.State, salckviews.UserEditModal, salckviews.SlackField)
		email := salckviews.GetInputText(event.View.State, salckviews.UserEditModal, salckviews.EmailField)
		githubName := salckviews.GetInputText(event.View.State, salckviews.UserEditModal, salckviews.GithubField)
		roles := salckviews.GetMultiSelectValues(event.View.State, salckviews.UserEditModal, salckviews.RolesField)
		teams := salckviews.GetMultiSelectValues(event.View.State, salckviews.UserEditModal, salckviews.TeamsField)

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
		userId, _ := slackflow.ParseEscapedLink(args[0])

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

	salckviews.GetUserEditModal(user)
	_, err := s.bag.Client.Slack.OpenViewContext(ctx, data.triggerId, salckviews.GetUserEditModal(user))

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
