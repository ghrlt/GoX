package subscriptions_service

import (
	"gox/database"
	"gox/database/models"

	"github.com/google/uuid"
)

func GetAll() ([]models.Subscription, error) {
	var subs []models.Subscription
	err := database.DB.Find(&subs, "is_accessible = ?", true).Error
	return subs, err
}

func GetByID(id uuid.UUID) (models.Subscription, error) {
	var sub models.Subscription
	err := database.DB.First(&sub, "id = ? AND is_accessible = ?", id, true).Error
	return sub, err
}
