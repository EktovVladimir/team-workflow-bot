package slackflow

import (
	"context"
	"team-workflow-bot/pkg/slackutils"

	"github.com/slack-go/slack"
)

type ShowOptionsModalParams struct {
	Title       string
	Text        string
	ConfirmText string
	CancelText  string
	MetaData    string
}

type Service struct {
	client *slack.Client
}

func NewService(client *slack.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) SendEphemeralErrorMessage(
	ctx context.Context,
	channel string,
	userId string,
	text string) {
	_, _, _, _ = s.client.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionPostEphemeral(userId),
		slack.MsgOptionAttachments(
			slack.Attachment{
				Color: "danger",
				Text:  text,
			},
		),
	)
}

func (s *Service) SendThreadEphemeralErrorMessage(
	ctx context.Context,
	channel string,
	ts string,
	userId string,
	text string) {
	_, _, _, _ = s.client.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionPostEphemeral(userId),
		slack.MsgOptionTS(ts),
		slack.MsgOptionAttachments(
			slack.Attachment{
				Color: "danger",
				Text:  text,
			},
		),
	)
}

func (s *Service) SendThreadErrorMessage(
	ctx context.Context,
	channel string,
	ts string,
	text string) {
	_, _, _, _ = s.client.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionTS(ts),
		slack.MsgOptionAttachments(
			slack.Attachment{
				Color: "danger",
				Text:  text,
			},
		),
	)
}

func (s *Service) GetMessageByTs(ctx context.Context, channelId string, ts string) (*slack.Message, error) {
	mess, err := s.client.GetConversationHistoryContext(ctx,
		&slack.GetConversationHistoryParameters{
			ChannelID: channelId,
			Inclusive: true,
			Latest:    ts,
			Limit:     1,
		})

	if err != nil {
		return nil, err
	}

	return &mess.Messages[0], nil
}

func (s *Service) ShowOptionsModal(
	ctx context.Context,
	triggerId string,
	callBackId string,
	params *ShowOptionsModalParams) (viewId string, err error) {
	blocks := []slack.Block{
		slackutils.GetMarkdownTextSectionBlock(params.Text),
	}

	viewRs, err := s.client.OpenViewContext(
		ctx,
		triggerId,
		slack.ModalViewRequest{
			CallbackID: callBackId,
			Type:       slack.VTModal,
			Title:      slackutils.GetEmojiPlainTextObject(params.Title),
			Submit:     slackutils.GetSimplePlainTextObject(params.ConfirmText),
			Close:      slackutils.GetSimplePlainTextObject(params.CancelText),
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
			PrivateMetadata: params.MetaData,
		})
	if err != nil {
		return "", err
	}
	return viewRs.View.ID, nil
}
