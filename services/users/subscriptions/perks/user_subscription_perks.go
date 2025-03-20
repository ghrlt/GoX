package user_sub_perks_service

import (
	"errors"

	"github.com/google/uuid"

	"gox/database"
	"gox/database/models"

	user_subscription_service "gox/services/users/subscriptions"
)

func AddPerks(userID uuid.UUID, userSubscriptionID uuid.UUID, collaborativeTeamCount int, maxProductsPerTeam int) error {
	userSubscription, err := user_subscription_service.Get(userID, userSubscriptionID)
	if err != nil {
		return err
	}
	if userSubscription == nil {
		return errors.New("user subscription not found")
	}

	subscriptionPerks := models.SubscriptionPerks{
		UserSubscriptionID:     userSubscriptionID,
		CollaborativeTeamCount: collaborativeTeamCount,
		MaxProductsPerTeam:     maxProductsPerTeam,
	}

	if err := database.DB.Create(&subscriptionPerks).Error; err != nil {
		return err
	}

	return nil
}

func GetPerks(userID uuid.UUID, userSubscriptionID uuid.UUID) (*models.SubscriptionPerks, error) {
	var subscriptionPerks models.SubscriptionPerks
	if err := database.DB.Where("user_subscription_id = ?", userSubscriptionID).First(&subscriptionPerks).Error; err != nil {
		return nil, err
	}
	return &subscriptionPerks, nil
}

func UpdatePerks(userID uuid.UUID, userSubscriptionID uuid.UUID, collaborativeTeamCount int, maxProductsPerTeam int) error {
	subscriptionPerks, err := GetPerks(userID, userSubscriptionID)
	if err != nil {
		return err
	}
	if subscriptionPerks == nil {
		return errors.New("subscription perks not found")
	}

	subscriptionPerks.CollaborativeTeamCount = collaborativeTeamCount
	subscriptionPerks.MaxProductsPerTeam = maxProductsPerTeam

	if err := database.DB.Save(&subscriptionPerks).Error; err != nil {
		return err
	}

	return nil
}
