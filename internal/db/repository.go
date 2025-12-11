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

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateUser(ctx context.Context, item *models.User) (string, error) {
	res, err := r.db.Collection("Users").InsertOne(ctx, item)
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

	_, err := r.db.Collection("Users").UpdateOne(
		ctx,
		bson.M{"slack_id": item.SlackId},
		bson.M{"$set": &updUser},
	)

	return err
}

func (r *Repository) GetBySlackId(ctx context.Context, slackId string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("Users").
		FindOne(ctx, bson.M{"slack_id": slackId}).
		Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, RecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetByGithubLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("Users").
		FindOne(ctx, bson.M{"github_login": login}).
		Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, RecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *Repository) CreateRole(ctx context.Context, item *models.Role) (string, error) {
	res, err := r.db.Collection("Roles").InsertOne(ctx, item)
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
	res, err := r.db.Collection("Teams").InsertOne(ctx, item)
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
	cursor, err := r.db.Collection("Roles").Find(ctx, bson.M{})
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
	cursor, err := r.db.Collection("Teams").Find(ctx, bson.M{})
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
