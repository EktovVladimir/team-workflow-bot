package models

import "fmt"

type ThreadRef struct {
	Key       string `bson:"key"`
	ChannelId string `bson:"channel_id"`
	Ts        string `bson:"ts"`
}

func NewThreadRef(channelId, ts string) *ThreadRef {
	return &ThreadRef{
		Key:       fmt.Sprintf("%s/%s", channelId, ts),
		ChannelId: channelId,
		Ts:        ts,
	}
}
