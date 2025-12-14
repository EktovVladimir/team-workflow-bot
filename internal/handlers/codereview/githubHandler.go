package codereview

import (
	"context"
	"errors"
	"fmt"
	"team-workflow-bot/internal/db"
	"team-workflow-bot/internal/integrations/githubflow"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/internal/models"
	"team-workflow-bot/pkg/slackutils"
	"time"

	"github.com/cbrgm/githubevents/v2/githubevents"
	"github.com/google/go-github/v79/github"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func (h *Handler) HandlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent) {
	action := event.GetAction()

	prInfo := githubflow.MapPullRequestInfoFromResponse(event.PullRequest)

	if action == githubevents.PullRequestEventOpenedAction ||
		action == githubevents.PullRequestEventReadyForReviewAction ||
		action == githubevents.PullRequestEventUnlabeledAction {

		h.handlePullRequestReady(ctx, event, prInfo)
	} else if action == githubevents.PullRequestEventClosedAction && prInfo.IsMerged {

		h.handlePullRequestMerged(ctx, event, prInfo)
	}
}

func (h *Handler) HandlePullRequestReviewEvent(ctx context.Context, event *github.PullRequestReviewEvent) {
	logrus.Infof("Handle PR review event: %s", event.GetAction())
}

func (h *Handler) handlePullRequestReady(ctx context.Context, event *github.PullRequestEvent, prInfo *models.PullRequestInfo) {
	// TODO реализовать
	// Автоматическое создание треда, если опция включена у юзера.
	// Проверка: пр открыт, не в драфте и нет спец. лейбла.
	// Сначала пытаемся найти существующий тред по ключу (номер задачи из ветки),
	// Если найден, и в треде не пролинкован этот ПР, то обновляем тред и БД.
	// Если не найден, создаем тред и сохраняем в БД.
}

func (h *Handler) handlePullRequestMerged(ctx context.Context, event *github.PullRequestEvent, prInfo *models.PullRequestInfo) {

	// TODO проверка на base ветку.
	// В разных репах, могут быть разные правила (develop или master, а может быть и для test веток тоже)
	// Пока ограничены условием в workflow файле, но если перейдем на вебхуки, то нужно будет это учитывать.

	dbCrThread, err := h.bag.DB.Repository.GetCodeReviewThreadByPullRequestRef(ctx, prInfo.Ref)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return
		}

		logrus.Errorf("Failed to get code review thread by PR ref: %v", err)
		return
	}

	oldPrInfo, updIndex, found := lo.FindIndexOf(dbCrThread.Context.PullRequests, func(item *models.PullRequestInfo) bool {
		return item.RefKey == prInfo.RefKey
	})
	if !found {
		logrus.Errorf("Failed to find PR info by PR ref: %v", prInfo.Ref)
		return
	}

	dbCrThread.Context.PullRequests[updIndex] = oldPrInfo

	threadRef := dbCrThread.Thread

	err = h.bag.Client.Slack.AddReactionContext(ctx, "white_check_mark", slack.ItemRef{
		Channel:   threadRef.ChannelId,
		Timestamp: threadRef.Ts,
	})
	if err != nil {
		logrus.Errorf("Failed to add reaction to code review thread: %v", err)
	}

	notificationText := fmt.Sprintf("Смерджено: %s", dbCrThread.Context.KeyedIssue)

	_, _, err = h.bag.Client.Slack.PostMessageContext(ctx, threadRef.ChannelId,
		slack.MsgOptionText(notificationText, false),
		slack.MsgOptionTS(threadRef.Ts),
		slack.MsgOptionBlocks(slackutils.GetMarkdownTextSectionBlock("Смерджено :white_check_mark:")))
	if err != nil {
		logrus.Errorf("Failed to post merged message to code review thread: %v", err)
		return
	}

	slackUserId := ""
	senderLogin := event.Sender.GetLogin()
	user, err := h.bag.DB.Repository.GetUserByGithubLogin(ctx, senderLogin)
	if err == nil {
		slackUserId = user.SlackId
	} else {
		slackUserId = dbCrThread.Context.Requester.SlackId
	}

	if slackUserId != "" {
		_, err = h.bag.Client.Slack.PostEphemeralContext(
			ctx,
			threadRef.ChannelId,
			slackUserId,
			slack.MsgOptionTS(threadRef.Ts),
			slack.MsgOptionBlocks(slackviews.GetDeployDatePickerBlocks(time.Now())...))
		if err != nil {
			logrus.Error("Failed to post deploy date picker: ", err)
		}
	}
}
