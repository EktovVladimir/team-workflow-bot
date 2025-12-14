package codereview

import (
	"context"
	"strings"
	"team-workflow-bot/internal/integrations/slackflow/slackviews"
	"team-workflow-bot/pkg/slackutils"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type DeployPickerSlackHandler struct{}

func NewDeployPickerSlackHandler() *DeployPickerSlackHandler {
	return &DeployPickerSlackHandler{}
}

func (d DeployPickerSlackHandler) HandleSlackAppMentionEvent(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, message *slackevents.AppMentionEvent) {
	client.Ack(*evt.Request)

	if !strings.Contains(message.Text, "/deploy-date") {
		return
	}

	deployDatePicker := slackviews.GetDeployDatePickerBlocks(time.Now())

	_, err := client.PostEphemeralContext(
		ctx,
		message.Channel,
		message.User,
		slack.MsgOptionTS(message.ThreadTimeStamp),
		slack.MsgOptionBlocks(deployDatePicker...))
	if err != nil {
		logrus.Error("Failed to post deploy date picker: ", err)
	}
}

func (d DeployPickerSlackHandler) HandleSlackBlockAction(ctx context.Context, evt *socketmode.Event, client *socketmode.Client, callback slack.InteractionCallback) {
	if len(callback.ActionCallback.BlockActions) == 0 {
		return
	}

	actionId := callback.ActionCallback.BlockActions[0].ActionID
	threadTs := callback.Container.ThreadTs
	channelId := callback.Channel.ID
	responseUrl := callback.ResponseURL

	if actionId == slackutils.GetActionId(slackviews.DeployDatePicker, slackviews.DeployDateField) {
		client.Ack(*evt.Request)
		return
	}

	if actionId == slackutils.GetActionId(slackviews.DeployDatePicker, slackviews.SubmitDeployDate) {
		client.Ack(*evt.Request)

		viewStateValues := callback.BlockActionState.Values

		deployDate := slackutils.GetDatePickerValue(viewStateValues, slackviews.DeployDatePicker, slackviews.DeployDateField)

		postBlocks := slackviews.GetDeployDatePostBlocks(deployDate)

		_, _, err := client.PostMessageContext(
			ctx,
			channelId,
			slack.MsgOptionBlocks(postBlocks...),
			slack.MsgOptionTS(threadTs),
		)
		if err != nil {
			logrus.Error("Failed to post deploy date message: ", err)
			return
		}

		_, _, _ = client.PostMessageContext(ctx, channelId, slack.MsgOptionDeleteOriginal(responseUrl))
	}
}
