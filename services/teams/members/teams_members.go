package team_member_service

import (
	"fmt"
	"gox/database"
	"gox/database/models"
	"strconv"

	"github.com/google/uuid"
)

func GetAll(teamID uuid.UUID) ([]models.TeamMember, error) {
	var members []models.TeamMember

	// Requête avec filtre : TeamID = teamID
	result := database.DB.Where("team_id = ?", teamID).Find(&members)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return nil, result.Error
	}

	return members, nil
}

func Get(teamID uuid.UUID, teamMemberID string) (models.TeamMember, error) {
	var member models.TeamMember

	memberIDInt, err := strconv.Atoi(teamMemberID)
	if err != nil {
		return models.TeamMember{}, fmt.Errorf("invalid team member ID: %v", err)
	}
	// Requête avec filtre : TeamID = teamID ET ID = teamMemberID
	result := database.DB.Where("team_id = ? AND id = ?", teamID, memberIDInt).First(&member)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return models.TeamMember{}, result.Error
	}

	return member, nil
}

func GetByMemberId(teamID, memberID uuid.UUID) (models.TeamMember, error) {
	var member models.TeamMember

	// Requête avec filtre : TeamID = teamID ET MemberID = memberID
	result := database.DB.Where("team_id = ? AND member_id = ?", teamID, memberID).First(&member)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return models.TeamMember{}, result.Error
	}

	return member, nil
}

func Add(teamID uuid.UUID, memberID uuid.UUID, role models.TeamMemberRole) error {
	// Vérification des champs requis
	if teamID == uuid.Nil || memberID == uuid.Nil || role == "" {
		return nil
	}

	// Création du TeamMember
	member := models.TeamMember{
		TeamID:   teamID,
		MemberID: memberID,
		Role:     role,
	}

	// Insertion en base
	if err := database.DB.Create(&member).Error; err != nil {
		return err
	}

	return nil
}

func Remove(teamID, memberID uuid.UUID) error {
	// Suppression du TeamMember
	result := database.DB.Where("team_id = ? AND member_id = ?", teamID, memberID).Delete(&models.TeamMember{})

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdateRole(teamID, memberID uuid.UUID, role models.TeamMemberRole) error {
	// Mise à jour du TeamMember
	result := database.DB.Model(&models.TeamMember{}).
		Where("team_id = ? AND member_id = ?", teamID, memberID).
		Update("role", role)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}
