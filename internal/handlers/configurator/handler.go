package configurator

import (
	"context"
	"errors"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/internal/models"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type interactivityData struct {
	channelId string
	userId    string
	triggerId string
}

type Handler struct {
	bag *bag.DependenciesBag
}

func NewSlackBotConfigurationHandler(
	bag *bag.DependenciesBag) *Handler {
	return &Handler{
		bag: bag,
	}
}

func (h *Handler) HandleSlackSlashCommand(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	cmd slack.SlashCommand) {
	if !h.isCommandApplicable(cmd) {
		return
	}

	args := strings.Split(cmd.Text, " ")

	if len(args) == 0 {
		return
	}

	user, err := h.bag.DB.Repository.GetBySlackId(ctx, cmd.UserID)
	if err != nil {
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(
			ctx,
			cmd.ChannelID,
			cmd.UserID,
			"Ошибка при получении данных о пользователе: "+err.Error())
		return
	}

	if user == nil || !user.HasRole(models.RoleAdmin) {
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(
			ctx,
			cmd.ChannelID,
			cmd.UserID,
			"У вас нет прав для использования этой команды.")
		return
	}

	data := interactivityData{
		channelId: cmd.ChannelID,
		userId:    cmd.UserID,
		triggerId: cmd.TriggerID,
	}

	client.Ack(*evt.Request)

	if len(args) >= 1 && args[0] == "user" {
		h.configureUsers(ctx, data, args[1:], true)
		return
	}
}

func (h *Handler) HandleSlackViewSubmission(
	ctx context.Context,
	evt *socketmode.Event,
	client *socketmode.Client,
	callback slack.InteractionCallback) {

	if callback.View.CallbackID == slackviews.GetCallbackId(slackviews.UserEditModal) {

		viewStateValues := callback.View.State.Values

		slackId := slackviews.GetSelectedUser(viewStateValues, slackviews.UserEditModal, slackviews.SlackField)
		email := slackviews.GetInputText(viewStateValues, slackviews.UserEditModal, slackviews.EmailField)
		githubName := slackviews.GetInputText(viewStateValues, slackviews.UserEditModal, slackviews.GithubField)
		roles := slackviews.GetMultiSelectValues(viewStateValues, slackviews.UserEditModal, slackviews.RolesField)
		teams := slackviews.GetMultiSelectValues(viewStateValues, slackviews.UserEditModal, slackviews.TeamsField)

		//TODO валидация

		client.Ack(*evt.Request)

		user, err := h.bag.DB.Repository.GetBySlackId(ctx, slackId)
		if err != nil && !errors.Is(err, db.ErrRecordNotFound) {
			logrus.Errorf("Error getting user by slack id %v: %v", slackId, err)
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
			_, err = h.bag.DB.Repository.CreateUser(ctx, user)
			if err != nil {
				logrus.Errorf("Error creating user with slack id %v: %v", slackId, err)
				return
			}
		} else {
			user.Email = email
			user.GitHubLogin = githubName
			user.Roles = roles
			user.Teams = teams

			err = h.bag.DB.Repository.UpdateUser(ctx, user)
			if err != nil {
				logrus.Errorf("Error updating user with slack id %v: %v", slackId, err)
				return
			}
		}
	}
}

func (h *Handler) isCommandApplicable(cmd slack.SlashCommand) bool {
	return cmd.Command == "/servit" && strings.HasPrefix(cmd.Text, "user")
}

func (h *Handler) configureUsers(
	ctx context.Context,
	data interactivityData,
	args []string,
	adminMode bool) {

	if len(args) == 0 {
		h.sendUserEditModal(ctx, data, &models.User{}, adminMode)
		return
	}

	if len(args) == 1 {
		userId, _, ok := slackflow.ParseEscapedLink(args[0])
		if !ok {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, data.channelId, data.userId,
				"Не удалось распознать пользователя из аргумента: "+args[0])
			return
		}

		user, err := h.bag.DB.Repository.GetBySlackId(ctx, userId)
		if err != nil && !errors.Is(err, db.ErrRecordNotFound) {
			h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(ctx, data.channelId, data.userId,
				"Ошибка при запросе БД: "+err.Error())
			return
		}

		if user == nil {
			user = &models.User{
				SlackId: userId,
			}
		}

		h.sendUserEditModal(ctx, data, user, adminMode)
		return
	}

	//TODO парсим аргументы из текста и создаем пользователя
}

func (h *Handler) sendUserEditModal(ctx context.Context, data interactivityData, user *models.User, adminMode bool) {

	_, err := h.bag.Client.Slack.OpenViewContext(
		ctx,
		data.triggerId,
		slackviews.GetUserEditModal(user, adminMode))

	if err != nil {
		h.bag.Services.Slack.SendSimpleEphemeralErrorMessage(
			ctx,
			data.channelId,
			data.userId,
			"Не удалось открыть модальное окно для редактирования пользователя. "+err.Error())
		return
	}
}
