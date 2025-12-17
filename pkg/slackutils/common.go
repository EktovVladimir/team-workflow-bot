package slackutils

import "fmt"

func GetEscapedMention(userId string) string {
	return fmt.Sprintf("<@%s>", userId)
}
