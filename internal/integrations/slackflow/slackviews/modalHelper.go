package slackviews

import (
	"team-workflow-bot/internal/models"
	"team-workflow-bot/pkg/slackutils"
)

func GetAvailableRolesOptionValues(roles []models.Role) []slackutils.SelectBlockOption {
	var options []slackutils.SelectBlockOption

	for _, role := range roles {
		options = append(options, slackutils.SelectBlockOption{
			First:  role.Name,
			Second: &role.Name,
			Third:  &role.Description,
		})
	}

	return options
}

func GetAvailableTeamsOptionValues(teams []models.Team) []slackutils.SelectBlockOption {
	var options []slackutils.SelectBlockOption

	for _, team := range teams {
		options = append(options, slackutils.SelectBlockOption{
			First:  team.Name,
			Second: &team.Name,
			Third:  &team.Description,
		})
	}

	return options
}
