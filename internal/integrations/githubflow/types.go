package githubflow

import (
	"context"

	"github.com/google/go-github/v79/github"
)

type GithubPullRequestEventHandler interface {
	HandlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent)
}

type GithubPullRequestReviewEventHandler interface {
	HandlePullRequestReviewEvent(ctx context.Context, event *github.PullRequestReviewEvent)
}

type GithubWorkflowRunEventHandler interface {
	HandleWorkflowRunEvent(ctx context.Context, event *github.WorkflowRunEvent)
}
