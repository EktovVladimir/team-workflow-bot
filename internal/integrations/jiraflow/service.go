package jiraflow

import (
	"context"
	"errors"
	"sync"
	"team-workflow-bot/internal/models"

	"github.com/andygrunwald/go-jira"
	"github.com/samber/lo"
)

type Service struct {
	client *jira.Client
}

func NewService(client *jira.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetIssueInfo(ctx context.Context, issueKey string) (*models.IssueInfo, error) {
	jIssue, _, err := s.client.Issue.GetWithContext(ctx, issueKey, nil)
	if err != nil {
		return nil, err
	}

	return &models.IssueInfo{
		Ref: &models.IssueRef{
			Key:     jIssue.Key,
			Project: jIssue.Fields.Project.Key,
			Owner:   "", // TODO Возможно нужно как-то определять владельца проекта
		},
		Title: jIssue.Fields.Summary,
	}, nil
}

func (s *Service) GetIssueInfoList(ctx context.Context, issueKeys []string) ([]*models.IssueInfo, error) {
	res := make([]*models.IssueInfo, 0)
	issueKeys = lo.Uniq(issueKeys)

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	errs := make([]error, 0)

	for _, key := range issueKeys {
		k := key
		wg.Add(1)
		go func(issueKey string) {
			defer wg.Done()
			issue, err := s.GetIssueInfo(ctx, issueKey)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			res = append(res, issue)
			mu.Unlock()
		}(k)
	}

	wg.Wait()

	if len(errs) > 0 {
		return res, errors.Join(errs...)
	}
	return res, nil
}

func (s *Service) GetIssueInfoListByRefs(ctx context.Context, refs []*models.IssueRef) ([]*models.IssueInfo, error) {
	issueKeys := lo.Map(refs, func(r *models.IssueRef, _ int) string {
		return r.Key
	})
	return s.GetIssueInfoList(ctx, issueKeys)
}
