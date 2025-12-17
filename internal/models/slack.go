package models

import (
	"fmt"
	"strings"
)

type MessageRef struct {
	Key       string `bson:"key"`
	ChannelId string `bson:"channel_id"`
	Ts        string `bson:"ts"`
}

func NewMessageRef(channelId, ts string) *MessageRef {
	return &MessageRef{
		Key:       fmt.Sprintf("%s/%s", channelId, ts),
		ChannelId: channelId,
		Ts:        ts,
	}
}

func ParseMessageRefFromKey(key string) *MessageRef {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) != 2 {
		return &MessageRef{}
	}
	return NewMessageRef(parts[0], parts[1])
}
