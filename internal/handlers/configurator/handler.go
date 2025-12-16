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
	"team-workflow-bot/pkg/slackutils"

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
	*bag.ServiceWithDependencies

	repository   *db.Repository
	slackClient  *slack.Client
	slackService *slackflow.Service
}

func NewSlackBotConfigurationHandler(b *bag.DependenciesBag) *Handler {
	return &Handler{
		ServiceWithDependencies: bag.NewServiceWithDependencies(b),
		repository:              b.DB.Repository,
		slackClient:             b.Client.Slack,
		slackService:            b.Services.Slack,
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

	user, err := h.repository.GetBySlackId(ctx, cmd.UserID)
	if err != nil {
		h.slackService.SendEphemeralErrorMessage(
			ctx,
			cmd.ChannelID,
			cmd.UserID,
			"Ошибка при получении данных о пользователе: "+err.Error())
		return
	}

	if user == nil || !user.HasRole(models.RoleAdmin) {
		h.slackService.SendEphemeralErrorMessage(
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

	if callback.View.CallbackID == slackutils.GetCallbackId(slackviews.UserEditModal) {

		viewStateValues := callback.View.State.Values

		slackId := slackutils.GetSelectedUser(viewStateValues, slackviews.UserEditModal, slackviews.SlackField)
		email := slackutils.GetInputText(viewStateValues, slackviews.UserEditModal, slackviews.EmailField)
		githubName := slackutils.GetInputText(viewStateValues, slackviews.UserEditModal, slackviews.GithubField)
		roles := slackutils.GetMultiSelectValues(viewStateValues, slackviews.UserEditModal, slackviews.RolesField)
		teams := slackutils.GetMultiSelectValues(viewStateValues, slackviews.UserEditModal, slackviews.TeamsField)

		//TODO валидация

		client.Ack(*evt.Request)

		user, err := h.repository.GetBySlackId(ctx, slackId)
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
			_, err = h.repository.CreateUser(ctx, user)
			if err != nil {
				logrus.Errorf("Error creating user with slack id %v: %v", slackId, err)
				return
			}
		} else {
			user.Email = email
			user.GitHubLogin = githubName
			user.Roles = roles
			user.Teams = teams

			err = h.repository.UpdateUser(ctx, user)
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
			h.slackService.SendEphemeralErrorMessage(ctx, data.channelId, data.userId,
				"Не удалось распознать пользователя из аргумента: "+args[0])
			return
		}

		user, err := h.repository.GetBySlackId(ctx, userId)
		if err != nil && !errors.Is(err, db.ErrRecordNotFound) {
			h.slackService.SendEphemeralErrorMessage(ctx, data.channelId, data.userId,
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

	_, err := h.slackClient.OpenViewContext(
		ctx,
		data.triggerId,
		slackviews.GetUserEditModal(user, adminMode))

	if err != nil {
		h.slackService.SendEphemeralErrorMessage(
			ctx,
			data.channelId,
			data.userId,
			"Не удалось открыть модальное окно для редактирования пользователя. "+err.Error())
		return
	}
}
