package githubflow

import (
	"context"
	"team-workflow-bot/internal/models"

	"github.com/google/go-github/v79/github"
)

type Service struct {
	client *github.Client
}

func NewService(client *github.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetAllCommits Получаем все модели коммитов из ПРа с помощью пагинации.
func (s *Service) GetAllCommits(ctx context.Context, prRef *models.PullRequestRef) ([]*models.CommitInfo, error) {
	gh := s.client

	res := make([]*models.CommitInfo, 0)
	page := 1
	pageSize := 100
	for {
		paginateOpt := &github.ListOptions{
			Page:    page,
			PerPage: pageSize,
		}

		commits, _, err := gh.PullRequests.ListCommits(ctx, prRef.Owner, prRef.Repo, prRef.Number, paginateOpt)
		if err != nil {
			return nil, err
		}

		if len(commits) == 0 {
			break
		}

		for _, commit := range commits {
			if commit.Commit == nil {
				continue
			}

			res = append(res, &models.CommitInfo{
				Message: commit.Commit.GetMessage(),
			})
		}

		page++
	}

	return res, nil
}
