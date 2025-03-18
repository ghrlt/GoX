package team_service

import (
	"fmt"
	"gox/database"
	"gox/database/models"
	"gox/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func IsUserInTeam(userID uuid.UUID, teamID uuid.UUID) (bool, error) {
	var count int64
	result := database.DB.Model(&models.TeamMember{}).Where("member_id = ? AND team_id = ?", userID, teamID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func IsUserInTeams(userID uuid.UUID, teamIDs []uuid.UUID) bool {
	var count int64
	result := database.DB.Model(&models.TeamMember{}).Where("member_id = ? AND team_id IN ?", userID, teamIDs).Count(&count)
	if result.Error != nil {
		utils.ConsoleLog("An error occured in IsUserInTeams: %v", result.Error).Error()
		return false
	}

	return count > 0
}

func Create(name string, teamType models.TeamType) (uuid.UUID, error) {
	// Vérification des champs requis
	if name == "" {
		return uuid.UUID{}, fmt.Errorf("name is required")
	}
	if teamType == "" {
		return uuid.UUID{}, fmt.Errorf("type is required")
	}

	// Création du Team
	team := models.Team{
		Name: name,
		Type: teamType,
	}

	// Insertion en base
	if err := database.DB.Create(&team).Error; err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating Team: %v", err)
	}

	// retourner le Team créé
	return team.ID, nil
}

func GetAll(db *gorm.DB) ([]models.Team, error) {
	var teams []models.Team

	// Récupération de tous les Teams
	result := database.DB.Find(&teams)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return nil, result.Error
	}

	return teams, nil
}

func Get(teamID uuid.UUID) (models.Team, error) {
	var team models.Team

	// Récupération du Team
	result := database.DB.Where("id = ?", teamID).First(&team)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return models.Team{}, result.Error
	}

	return team, nil
}

func GetPersonalTeamByMemberID(memberID uuid.UUID) (models.Team, error) {
	var team models.Team
	result := database.DB.Where("type = ? AND id = (SELECT team_id FROM team_members WHERE member_id = ?)", models.TeamTypePersonal, memberID).First(&team)
	if result.Error != nil {
		return models.Team{}, result.Error
	}

	return team, nil
}

func GetTeamsByMemberID(memberID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	result := database.DB.Where("id IN (SELECT team_id FROM team_members WHERE member_id = ?)", memberID).Find(&teams)
	if result.Error != nil {
		return nil, result.Error
	}

	return teams, nil
}

func UpdateName(teamID uuid.UUID, name string) error {
	// Mise à jour du Team
	result := database.DB.Model(&models.Team{}).
		Where("id = ?", teamID).
		Updates(map[string]interface{}{
			"name": name,
		})

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func Delete(teamID uuid.UUID) error {
	// Suppression du Team
	result := database.DB.Where("id = ?", teamID).Delete(&models.Team{})

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}
