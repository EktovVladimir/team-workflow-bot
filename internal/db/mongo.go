package db

import (
	"context"
	"team-workflow-bot/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.Mongo.Connection)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	usersColl := db.Collection(UsersCollection)
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slack_id", Value: 1}},
			Options: options.Index().SetName("idx_users_slack_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "github_login", Value: 1}},
			Options: options.Index().SetName("idx_users_github_login"),
		},
	}
	if _, err := usersColl.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return err
	}

	crtColl := db.Collection(CodeReviewThreadsCollection)
	crtIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "context.keyed_issue", Value: 1}},
			Options: options.Index().SetName("idx_crt_context_keyed_issue"),
		},
		{
			Keys:    bson.D{{Key: "context.pull_requests.ref.owner", Value: 1}},
			Options: options.Index().SetName("idx_crt_pr_ref_owner"),
		},
		{
			Keys:    bson.D{{Key: "context.pull_requests.ref.repo", Value: 1}},
			Options: options.Index().SetName("idx_crt_pr_ref_repo"),
		},
		{
			Keys:    bson.D{{Key: "context.pull_requests.ref.number", Value: 1}},
			Options: options.Index().SetName("idx_crt_pr_ref_number"),
		},
		{
			Keys:    bson.D{{Key: "context.pull_requests.ref.key", Value: 1}},
			Options: options.Index().SetName("idx_crt_pr_ref_key"),
		},
		{
			Keys:    bson.D{{Key: "thread.key", Value: 1}},
			Options: options.Index().SetName("idx_crt_thread_ref_key"),
		},
	}
	if _, err := crtColl.Indexes().CreateMany(ctx, crtIndexes); err != nil {
		return err
	}

	return nil
}
