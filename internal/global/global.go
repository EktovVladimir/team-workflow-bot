package global

import (
	"context"
	"team-workflow-bot/internal/db"
)

//TODO кэширование

type Storage struct {
	AvailableRoles []db.Role
	AvailableTeams []db.Team
}

var storage *Storage

func GetStorage() *Storage {
	return storage
}

func InitGlobalStorageData(ctx context.Context, repo *db.Repository) error {
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
