package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"team-workflow-bot/internal/bag"
	"team-workflow-bot/internal/integrations/githubflow"
	"team-workflow-bot/internal/models"

	"sync"

	"github.com/google/go-github/v79/github"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var ErrNoPRsOrIssues = errors.New("no pull requests or issues to retrieve")
var ErrDbUserNotFoundButOtherData = errors.New("db user not found but other data exists")

var ErrUserRefIsEmpty = errors.New("user reference is empty")
var ErrGithubUserPublicEmailNotFound = errors.New("github user has no public email")
var ErrSlackUserPublicEmailNotFound = errors.New("slack user has no public email")
var ErrNothingToCollect = errors.New("nothing to collect")

type Retriever struct {
	bag *bag.DependenciesBag
}

func NewRetriever(bag *bag.DependenciesBag) *Retriever {
	return &Retriever{
		bag: bag,
	}
}

func (r *Retriever) CollectCodeReviewContextFromSlack(ctx context.Context, request *models.CodeReviewCollectRequest) (*models.CodeReviewContext, error) {
	issueKeys := request.GetIssueKeys()

	res := &models.CodeReviewContext{
		Requester: request.Requester,
		Reviewers: request.Reviewers,
	}

	prInfos, err := r.getPullRequestInfoList(ctx, request.PullRequests...)
	if err != nil {
		return res, err
	}

	res.PullRequests = prInfos

	if len(prInfos) != 0 {
		if !request.DisableCollectReviewersFromPr {
			reviewerGhLogins := lo.Uniq(lo.FlatMap(prInfos, func(pr *models.PullRequestInfo, _ int) []string {
				return pr.Reviewers
			}))
			dbReviewers := r.GetUserListByGithubLoginsSafe(ctx, reviewerGhLogins)
			reviewersFromGhRefs := lo.Map(dbReviewers, func(u *models.User, _ int) *models.UserRef {
				return u.ToRef()
			})

			//Добавляем ревьюверов к общему списку и удаляем дубликаты по slackId,
			//но оставляем githubLogin, если slackId нет
			res.Reviewers = append(res.Reviewers, reviewersFromGhRefs...)
			res.Reviewers = lo.UniqBy(res.Reviewers, func(r *models.UserRef) string {
				if r.SlackId != "" {
					return r.SlackId
				}
				return r.GithubLogin
			})
		}

		headBrunchNames := lo.Map(prInfos, func(pr *models.PullRequestInfo, _ int) string {
			return pr.HeadBranch
		})
		issueKeysFromBranches := getUniqKeysFromMessages(headBrunchNames...)

		// Нужно задать некий ключ код-ревью, по которому сможем объединить ПРы из разных репозиториев.
		// И по этому ключу потом искать существующее код-ревью в БД.
		// Если можем, используем номер первой задачи из названия ветки.
		// Если ветка не содержит номер, то просто используем название ветки.
		if len(issueKeysFromBranches) != 0 {
			res.KeyedIssue = issueKeysFromBranches[0]
		} else {
			res.KeyedIssue = headBrunchNames[0]
		}

		if !request.DisableCollectIssuesFromPr {
			issueKeysFromCommits := r.GetIssueKeysFromCommitMessages(ctx, request.PullRequests...)

			//Добавляем задачи к общему списку и удаляем дубликаты
			issueKeys = lo.Union(issueKeys, issueKeysFromBranches, issueKeysFromCommits)
		}
	}

	issues, err := r.bag.Services.Jira.GetIssueInfoList(ctx, issueKeys)
	if err != nil {
		return res, err
	}

	res.Issues = issues

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

	dbUser, err := rep.GetUserByGithubLogin(ctx, userRef.GithubLogin)

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
	uniqPrRefs := getUniqPullRequestRefs(prRefs...)

	messages := make([]string, 0, len(uniqPrRefs))

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, prRef := range uniqPrRefs {
		wg.Add(1)
		go func(ref *models.PullRequestRef) {
			defer wg.Done()
			ghCommits, err := r.bag.Services.Github.GetAllCommits(ctx, ref)
			if err != nil {
				logrus.Warning("Ignoring error while retrieving commits for PR:", err)
				return
			}

			msgs := lo.Map(ghCommits, func(c *models.CommitInfo, _ int) string {
				return c.Message
			})
			mu.Lock()
			messages = append(messages, msgs...)
			mu.Unlock()
		}(prRef)
	}

	wg.Wait()

	return getUniqKeysFromMessages(messages...)
}

func (r *Retriever) getPullRequestInfoList(ctx context.Context, prRefs ...*models.PullRequestRef) ([]*models.PullRequestInfo, error) {
	uniqPrRefs := getUniqPullRequestRefs(prRefs...)
	prInfos := make([]*models.PullRequestInfo, 0, len(uniqPrRefs))

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, prRef := range uniqPrRefs {
		wg.Add(1)
		go func(ref *models.PullRequestRef) {
			defer wg.Done()
			ghPr, _, err := r.bag.Client.GitHub.PullRequests.Get(ctx, ref.Owner, ref.Repo, ref.Number)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			prInfo := githubflow.MapPullRequestInfoFromResponse(ghPr)
			mu.Lock()
			prInfos = append(prInfos, prInfo)
			mu.Unlock()
		}(prRef)
	}

	wg.Wait()

	if len(errs) > 0 {
		return prInfos, errors.Join(errs...)
	}
	return prInfos, nil
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

func getUniqPullRequestRefs(prRefs ...*models.PullRequestRef) []*models.PullRequestRef {
	return lo.UniqBy(prRefs, func(r *models.PullRequestRef) string {
		return r.ToKey()
	})
}
