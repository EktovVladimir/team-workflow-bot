package codereview

import (
	"strings"
	"team-workflow-bot/internal/integrations/slackflow"
	"team-workflow-bot/internal/models"

	"github.com/samber/lo"
)

func getParsedFromUrlPullRequestRefs(prUrls []string) []*models.PullRequestRef {
	uniqUrls := lo.Uniq(prUrls)
	return lo.FilterMap(uniqUrls, func(arg string, i int) (*models.PullRequestRef, bool) {
		if strings.ContainsAny(arg, "<|>") {
			url, _, ok := slackflow.ParseEscapedLink(arg)
			if ok {
				arg = url
			}
		}

		return models.ParsePullRequestRefFromUrl(arg)
	})
}

func getParsedFromUrlIssueRefs(issueUrls []string) []*models.IssueRef {
	uniqUrls := lo.Uniq(issueUrls)
	return lo.FilterMap(uniqUrls, func(arg string, i int) (*models.IssueRef, bool) {
		if strings.ContainsAny(arg, "<|>") {
			url, _, ok := slackflow.ParseEscapedLink(arg)
			if ok {
				arg = url
			}
		}

		return models.ParseIssueRefFromUrl(arg)
	})
}

func getParsedFromFormatUserRefs(slackUserMention []string) []*models.UserRef {
	uniqMentions := lo.Uniq(slackUserMention)
	return lo.FilterMap(uniqMentions, func(arg string, i int) (*models.UserRef, bool) {
		userId, _, ok := slackflow.ParseEscapedLink(arg)
		if !ok || !strings.HasPrefix(userId, "U") {
			return nil, false
		}
		return &models.UserRef{
			SlackId: userId,
		}, true
	})
}
