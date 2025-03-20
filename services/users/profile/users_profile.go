package user_profile_service

import (
	"gox/database"
	"gox/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Create(db *gorm.DB, userID uuid.UUID, profile models.UserProfile) error {
	// Création de l'utilisateur
	result := database.DB.Create(&models.UserProfile{
		CustomerID: userID,
		Username:   profile.Username,
		// AvatarURL:          profile.AvatarURL,
		PublicStatsDisplay: profile.PublicStatsDisplay,
	})

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func Get(db *gorm.DB, userID uuid.UUID) (models.UserProfile, error) {
	// Récupération de l'utilisateur
	var user models.UserProfile
	result := database.DB.Where("user_id = ?", userID).First(&user)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return models.UserProfile{}, result.Error
	}

	return user, nil
}

func Update(db *gorm.DB, userID uuid.UUID, profile models.UserProfile) error {
	// Mise à jour de l'utilisateur
	result := database.DB.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(profile)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func Delete(db *gorm.DB, userID uuid.UUID) error {
	// Suppression de l'utilisateur
	result := database.DB.Where("user_id = ?", userID).Delete(&models.UserProfile{})

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}
