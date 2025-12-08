package slackflow

import "regexp"

// ParseEscapedLink парсит записи вида:
// "<url|text>"
// <@U12345678|username>
// <#C12345678|channelname>
// Возвращает url/id и текст/имя
func ParseEscapedLink(text string) (string, string) {
	regex := `<[@#]?([^|]+)\|?([^>]*)>`
	re := regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		return matches[1], matches[2]
	}
	return "", ""

}
