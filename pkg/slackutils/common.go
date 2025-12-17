package slackutils

import (
	"fmt"
	"time"
)

func GetEscapedMention(userId string) string {
	return fmt.Sprintf("<@%s>", userId)
}

func GetTimeStamp(t time.Time) string {
	sec := t.Unix()
	usec := t.Nanosecond() / 1000
	return fmt.Sprintf("%d.%06d", sec, usec)
}

func GetTimeStampShort(t time.Time) string {
	sec := t.Unix()
	return fmt.Sprintf("%d", sec)
}
