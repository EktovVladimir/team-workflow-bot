package global

import (
	"context"
	"team-workflow-bot/internal/models"
)

//TODO кэширование

type repository interface {
	GetAllRoles(ctx context.Context) ([]models.Role, error)
	GetAllTeams(ctx context.Context) ([]models.Team, error)
}

type Storage struct {
	AvailableRoles []models.Role
	AvailableTeams []models.Team
}

var storage *Storage

func GetStorage() *Storage {
	return storage
}

func InitGlobalStorageData(ctx context.Context, repo repository) error {
	roles, err := repo.GetAllRoles(ctx)
	if err != nil {
		return err
	}

	teams, err := repo.GetAllTeams(ctx)
	if err != nil {
		return err
	}

	storage = &Storage{
		AvailableRoles: roles,
		AvailableTeams: teams,
	}

	return nil
}
