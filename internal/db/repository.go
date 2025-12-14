package db

import (
	"context"
	"errors"
	"team-workflow-bot/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	db *mongo.Database
}

const (
	UsersCollection             = "Users"
	RolesCollection             = "Roles"
	TeamsCollection             = "Teams"
	CodeReviewThreadsCollection = "code_review_threads"
)

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateUser(ctx context.Context, item *models.User) (string, error) {
	res, err := r.db.Collection(UsersCollection).InsertOne(ctx, item)
	if err != nil {
		return "", err
	}

	switch v := res.InsertedID.(type) {
	case string:
		return v, nil
	default:
		return "", nil
	}
}

func (r *Repository) UpdateUser(ctx context.Context, item *models.User) error {

	updUser := *item
	updUser.Id = ""

	//TODO обновлять по ID
	_, err := r.db.Collection(UsersCollection).UpdateOne(
		ctx,
		bson.M{"slack_id": item.SlackId},
		bson.M{"$set": &updUser},
	)

	return err
}

func (r *Repository) GetBySlackId(ctx context.Context, slackId string) (*models.User, error) {
	var user models.User
	err := r.db.Collection(UsersCollection).
		FindOne(ctx, bson.M{"slack_id": slackId}).
		Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetByGithubLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.Collection(UsersCollection).
		FindOne(ctx, bson.M{"github_login": login}).
		Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *Repository) CreateRole(ctx context.Context, item *models.Role) (string, error) {
	res, err := r.db.Collection(RolesCollection).InsertOne(ctx, item)
	if err != nil {
		return "", err
	}

	switch v := res.InsertedID.(type) {
	case string:
		return v, nil
	default:
		return "", nil
	}
}

func (r *Repository) CreateTeam(ctx context.Context, item *models.Team) (string, error) {
	res, err := r.db.Collection(TeamsCollection).InsertOne(ctx, item)
	if err != nil {
		return "", err
	}

	switch v := res.InsertedID.(type) {
	case string:
		return v, nil
	default:
		return "", nil
	}
}

func (r *Repository) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	cursor, err := r.db.Collection(RolesCollection).Find(ctx, bson.M{})
	if err != nil {
		return []models.Role{}, err
	}
	if err = cursor.All(ctx, &roles); err != nil {
		return []models.Role{}, err
	}

	if roles == nil {
		roles = []models.Role{}
	}

	return roles, nil
}

func (r *Repository) GetAllTeams(ctx context.Context) ([]models.Team, error) {
	var teams []models.Team
	cursor, err := r.db.Collection(TeamsCollection).Find(ctx, bson.M{})
	if err != nil {
		return []models.Team{}, err
	}
	if err = cursor.All(ctx, &teams); err != nil {
		return []models.Team{}, err
	}

	if teams == nil {
		teams = []models.Team{}
	}

	return teams, nil
}

func (r *Repository) CreateCodeReviewThread(ctx context.Context, item *models.CodeReviewThread) (string, error) {
	res, err := r.db.Collection(CodeReviewThreadsCollection).InsertOne(ctx, item)
	if err != nil {
		return "", err
	}

	switch v := res.InsertedID.(type) {
	case string:
		return v, nil
	default:
		return "", nil
	}
}

func (r *Repository) GetCodeReviewThreadByContextKey(ctx context.Context, key string) (*models.CodeReviewThread, error) {
	var crt models.CodeReviewThread
	err := r.db.Collection(CodeReviewThreadsCollection).
		FindOne(ctx, bson.M{"context.key": key}).
		Decode(&crt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &crt, nil
}

func (r *Repository) GetCodeReviewThreadByPullRequestRef(ctx context.Context, ref *models.PullRequestRef) (*models.CodeReviewThread, error) {
	var crt models.CodeReviewThread
	refKey := ref.ToKey()
	err := r.db.Collection(CodeReviewThreadsCollection).
		FindOne(ctx, bson.M{"context.pull_requests.ref_key": refKey}).
		Decode(&crt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &crt, nil
}

func (r *Repository) UpdateCodeReviewThread(ctx context.Context, item *models.CodeReviewThread) error {

	updCrt := *item
	updCrt.Id = ""

	//TODO обновлять по ID
	_, err := r.db.Collection(CodeReviewThreadsCollection).UpdateOne(
		ctx,
		bson.M{"context.key": item.Context.KeyedIssue},
		bson.M{"$set": &updCrt},
	)

	return err
}
