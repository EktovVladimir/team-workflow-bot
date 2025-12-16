package slackflow

import (
	"context"

	"github.com/slack-go/slack"
)

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
