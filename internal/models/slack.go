package models

type ThreadRef struct {
	ChannelId string `bson:"channel_id"`
	Ts        string `bson:"ts"`
}
