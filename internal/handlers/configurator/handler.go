package configurator

import (
	"context"
	"errors"
	"log"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
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

	if event.View.CallbackID == slackviews.GetCallbackId(slackviews.UserEditModal) {

		slackId := slackviews.GetSelectedUser(event.View.State, slackviews.UserEditModal, slackviews.SlackField)
		email := slackviews.GetInputText(event.View.State, slackviews.UserEditModal, slackviews.EmailField)
		githubName := slackviews.GetInputText(event.View.State, slackviews.UserEditModal, slackviews.GithubField)
		roles := slackviews.GetMultiSelectValues(event.View.State, slackviews.UserEditModal, slackviews.RolesField)
		teams := slackviews.GetMultiSelectValues(event.View.State, slackviews.UserEditModal, slackviews.TeamsField)

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

	slackviews.GetUserEditModal(user)
	_, err := s.bag.Client.Slack.OpenViewContext(ctx, data.triggerId, slackviews.GetUserEditModal(user))

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
