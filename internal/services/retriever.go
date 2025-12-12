package services

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/integrations/githubflow"
	"team-workflow-bot/internal/models"

	"github.com/google/go-github/v79/github"
	"github.com/samber/lo"
	"github.com/slack-go/slack"
)

var ErrNoPRsOrIssues = errors.New("no pull requests or issues to retrieve")
var ErrDbUserNotFoundButOtherData = errors.New("db user not found but other data exists")

var ErrUserRefIsEmpty = errors.New("user reference is empty")
var ErrGithubUserPublicEmailNotFound = errors.New("github user has no public email")
var ErrSlackUserPublicEmailNotFound = errors.New("slack user has no public email")

type Retriever struct {
	bag *bag.DependenciesBag
}

func NewRetriever(bag *bag.DependenciesBag) *Retriever {
	return &Retriever{
		bag: bag,
	}
}

func (r *Retriever) CollectCodeReviewContextFromSlack(ctx context.Context, request *models.RequestRef) (*models.CodeReviewContext, error) {

	//TODO Пока что работаем через ПРы.
	if len(request.PullRequests) == 0 /* && len(request.Issues) == 0 */ {
		return nil, ErrNoPRsOrIssues
	}

	gh := r.bag.Client.GitHub

	dbRequester, _ := r.GetUserBySlackIdSafe(ctx, request.Requester)

	prRefs := getUniqPullRequestRefs(request.PullRequests...)
	prInfos := make([]*models.PullRequestInfo, 0)
	for _, prRef := range prRefs {
		ghPr, _, err := gh.PullRequests.Get(ctx, prRef.Owner, prRef.Repo, prRef.Number)
		if err != nil {
			return nil, err
		}

		prInfo := githubflow.MapPullRequestInfoFromResponse(ghPr)
		prInfos = append(prInfos, prInfo)
	}

	//Собираем ревьюверов из ПРов
	reviewerGhLogins := lo.FlatMap(prInfos, func(pr *models.PullRequestInfo, _ int) []string {
		return pr.Reviewers
	})

	//Ревьюверы из запроса
	reviewerGhLogins = append(reviewerGhLogins, lo.Map(request.Reviewers, func(r *models.UserRef, _ int) string {
		return r.GithubLogin
	})...)

	reviewerGhLogins = lo.Uniq(reviewerGhLogins)

	dbReviewers := r.GetUserListByGithubLoginsSafe(ctx, reviewerGhLogins)
	reviewerRefs := lo.Map(dbReviewers, func(u *models.User, _ int) *models.UserRef {
		return u.ToRef()
	})

	headBrunchNames := lo.Map(prInfos, func(pr *models.PullRequestInfo, _ int) string {
		return pr.HeadBranch
	})

	mainIssueKeys := getUniqKeysFromMessages(headBrunchNames...)
	issueKeys := r.GetIssueKeysFromCommitMessages(ctx, request.PullRequests...)

	issues, err := r.bag.Services.Jira.GetIssueInfoList(ctx, lo.Union(mainIssueKeys, issueKeys))
	if err != nil {
		return nil, err
	}

	res := &models.CodeReviewContext{
		Requester: &models.UserRef{
			GithubLogin: dbRequester.GitHubLogin,
			SlackId:     dbRequester.SlackId,
		},
		Reviewers:    reviewerRefs,
		PullRequests: prInfos,
		Issues:       issues,
	}

	return res, nil
}

func (r *Retriever) CollectCodeReviewContextFromSlackLite(ctx context.Context, request *models.RequestRef) (*models.CodeReviewContext, error) {

	if len(request.PullRequests) == 0 {
		return nil, ErrNoPRsOrIssues
	}

	gh := r.bag.Client.GitHub

	prRefs := getUniqPullRequestRefs(request.PullRequests...)
	prInfos := make([]*models.PullRequestInfo, 0)
	for _, prRef := range prRefs {
		ghPr, _, err := gh.PullRequests.Get(ctx, prRef.Owner, prRef.Repo, prRef.Number)
		if err != nil {
			return nil, err
		}

		prInfo := githubflow.MapPullRequestInfoFromResponse(ghPr)
		prInfos = append(prInfos, prInfo)
	}

	issueKeys := lo.Map(request.Issues, func(issueRef *models.IssueRef, _ int) string {
		return issueRef.Number
	})

	issues, err := r.bag.Services.Jira.GetIssueInfoList(ctx, issueKeys)
	if err != nil {
		return nil, err
	}

	res := &models.CodeReviewContext{
		Requester:    request.Requester,
		Reviewers:    request.Reviewers,
		PullRequests: prInfos,
		Issues:       issues,
	}

	return res, nil
}

// GetUserByGithubLoginSafe Пытаемся найти пользователя по GitHub login.
// Важно! Ожидается что в userRef обязательно передан GithubLogin.
// Если пользователь не найден в БД, пытаемся вытащить его Slack ID через GitHub->email->Slack.
// Если пользователь не найден в БД, всё равно возвращаем структуру пользователя, и ошибку ErrDbUserNotFoundButOtherData.
// Иные ошибки игнорируются.
func (r *Retriever) GetUserByGithubLoginSafe(ctx context.Context, userRef *models.UserRef) (*models.User, error) {
	rep := r.bag.DB.Repository

	if userRef == nil || userRef.GithubLogin == "" {
		return &models.User{}, ErrUserRefIsEmpty
	}

	dbUser, err := rep.GetByGithubLogin(ctx, userRef.GithubLogin)

	if err == nil {
		return dbUser, nil
	}

	if userRef.SlackId != "" {
		return &models.User{
			GitHubLogin: userRef.GithubLogin,
			SlackId:     userRef.SlackId,
		}, ErrDbUserNotFoundButOtherData
	}

	slUser, err := r.GetSlackUserByGithubLogin(ctx, userRef.GithubLogin)
	if err != nil {
		return &models.User{
			GitHubLogin: userRef.GithubLogin,
		}, ErrDbUserNotFoundButOtherData
	}

	return &models.User{
		SlackId:     slUser.ID,
		GitHubLogin: userRef.GithubLogin,
	}, ErrDbUserNotFoundButOtherData
}

// GetUserBySlackIdSafe Пытаемся найти пользователя по Slack ID.
// Важно! Ожидается что в userRef обязательно передан SlackId.
// Если пользователь не найден в БД, пытаемся вытащить его github login через Slack->email->GitHub.
// Если пользователь не найден в БД, всё равно возвращаем структуру пользователя, и ошибку ErrDbUserNotFoundButOtherData.
// Иные ошибки игнорируются.
func (r *Retriever) GetUserBySlackIdSafe(ctx context.Context, userRef *models.UserRef) (*models.User, error) {
	rep := r.bag.DB.Repository

	if userRef == nil || userRef.SlackId == "" {
		return &models.User{}, ErrUserRefIsEmpty
	}

	dbUser, err := rep.GetBySlackId(ctx, userRef.SlackId)

	if err == nil {
		return dbUser, nil
	}

	if userRef.GithubLogin != "" {
		return &models.User{
			SlackId:     userRef.SlackId,
			GitHubLogin: userRef.GithubLogin,
		}, ErrDbUserNotFoundButOtherData
	}

	ghUser, err := r.GetGithubUserBySlackId(ctx, userRef.SlackId)
	if err != nil {
		return &models.User{
			SlackId: userRef.SlackId,
		}, ErrDbUserNotFoundButOtherData
	}

	return &models.User{
		SlackId:     userRef.SlackId,
		GitHubLogin: ghUser.GetLogin(),
	}, ErrDbUserNotFoundButOtherData
}

// GetUserListByGithubLoginsSafe Пытаемся найти пользователей по списку GitHub логинов.
// Если пользователь не найден в БД, пытаемся вытащить его Slack ID через GitHub->email->Slack.
// Любые ошибки игнорируются. Возвращаем то, что удалось найти.
func (r *Retriever) GetUserListByGithubLoginsSafe(ctx context.Context, githubLogins []string) []*models.User {
	res := make([]*models.User, 0)
	for _, ghLogin := range githubLogins {
		user, _ := r.GetUserByGithubLoginSafe(ctx, &models.UserRef{GithubLogin: ghLogin})
		res = append(res, user)
	}
	return res
}

func (r *Retriever) GetSlackUserByGithubLogin(ctx context.Context, githubLogin string) (*slack.User, error) {
	sl := r.bag.Client.Slack
	gh := r.bag.Client.GitHub

	rs, _, err := gh.Users.Get(ctx, githubLogin)
	if err != nil {
		return nil, err
	}

	email := rs.GetEmail()
	if email == "" {
		return nil, ErrGithubUserPublicEmailNotFound
	}

	slUsers, err := sl.GetUserByEmailContext(ctx, email)
	if err != nil {
		return nil, err
	}

	return slUsers, nil
}

func (r *Retriever) GetGithubUserBySlackId(ctx context.Context, slackId string) (*github.User, error) {
	sl := r.bag.Client.Slack
	gh := r.bag.Client.GitHub

	slUser, err := sl.GetUserInfoContext(ctx, slackId)
	if err != nil {
		return nil, err
	}

	email := slUser.Profile.Email
	if email == "" {
		return nil, ErrSlackUserPublicEmailNotFound
	}

	ghUsers, _, err := gh.Search.Users(ctx, email, &github.SearchOptions{})
	if err != nil {
		return nil, err
	}

	if ghUsers.GetTotal() == 0 {
		return nil, ErrGithubUserPublicEmailNotFound
	}

	return ghUsers.Users[0], nil
}

func (r *Retriever) GetIssueKeysFromCommitMessages(ctx context.Context, prRefs ...*models.PullRequestRef) []string {
	messages := make([]string, 0)
	for _, prRef := range prRefs {
		ghCommits, err := r.bag.Services.Github.GetAllCommits(ctx, prRef)
		if err != nil {
			log.Println("Ignoring error while retrieving commits for PR:", err)
			continue
		}

		messages = append(messages, lo.Map(ghCommits, func(c *models.CommitInfo, i int) string {
			return c.Message
		})...)
	}

	return getUniqKeysFromMessages(messages...)
}

// TODO в утилиты
func getUniqKeysFromMessages(messages ...string) []string {
	rxp := regexp.MustCompile(`(?i)(?:\s|/|^)(?P<issue>[A-Za-z]+-[0-9]+)(?:\s|[-_]|$)`)

	keys := lo.FilterMap(messages, func(s string, i int) (string, bool) {
		m := rxp.FindStringSubmatch(s)
		if len(m) != 2 {
			return "", false
		}

		key := m[1]

		key = strings.Trim(key, " \n\t ")
		key = strings.ToUpper(key)

		return key, true
	})

	return lo.Uniq(keys)
}

func getUniqPullRequestRefs(prRefs ...*models.PullRequestRef) []models.PullRequestRef {
	return lo.Uniq(lo.Map(prRefs, func(pr *models.PullRequestRef, _ int) models.PullRequestRef {
		return *pr
	}))
}
