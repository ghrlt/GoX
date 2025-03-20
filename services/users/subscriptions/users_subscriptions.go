package user_subscription_service

import (
	"errors"
	"gox/database"
	"gox/database/models"
	"time"

	"github.com/google/uuid"
)

func GetAll(userID uuid.UUID) ([]models.UserSubscription, error) {
	var subscriptions []models.UserSubscription
	if err := database.DB.Where("customer_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func Get(userID uuid.UUID, subscriptionID uuid.UUID) (*models.UserSubscription, error) {
	var subscription models.UserSubscription
	if err := database.DB.Where("customer_id = ? AND id = ?", userID, subscriptionID).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

func GetActive(userID uuid.UUID) (*models.UserSubscription, error) {
	subscriptions, err := GetAll(userID)
	if err != nil {
		return nil, err
	}

	for _, subscription := range subscriptions {
		if subscription.StartAt.Before(time.Now()) && subscription.StartAt.AddDate(0, 0, subscription.Subscription.ValidForInDays).After(time.Now()) {
			return &subscription, nil
		}
	}

	return nil, nil
}

func GetInactives(userID uuid.UUID) ([]models.UserSubscription, error) {
	subscriptions, err := GetAll(userID)
	if err != nil {
		return nil, err
	}

	var inactiveSubscriptions []models.UserSubscription
	for _, subscription := range subscriptions {
		if subscription.StartAt.AddDate(0, 0, subscription.Subscription.ValidForInDays).Before(time.Now()) {
			inactiveSubscriptions = append(inactiveSubscriptions, subscription)
		}
	}

	return inactiveSubscriptions, nil
}

func Create(userID uuid.UUID, subscriptionID uuid.UUID, autoRenew bool) (*models.UserSubscription, error) {
	var subscription models.Subscription
	if err := database.DB.Where("id = ?", subscriptionID).First(&subscription).Error; err != nil {
		return nil, err
	}

	var userSubscription models.UserSubscription
	// ~ Check if user already has an active subscription
	activeSubscription, err := GetActive(userID)
	if err != nil {
		return nil, err
	}
	if activeSubscription != nil {
		// ~ Return error if user already has an active subscription
		return nil, errors.New("user already has an active subscription")
	}

	userSubscription = models.UserSubscription{
		CustomerID:     userID,
		SubscriptionID: subscriptionID,
		AutoRenew:      autoRenew,
		StartAt:        time.Now(),
		TotalPrice:     subscription.Price,
		IsAccessible:   false, // ~ Subscription is not accessible until perks are added
	}
	if err := database.DB.Create(&subscription).Error; err != nil {
		return nil, err
	}

	return &userSubscription, nil
}

func Update(userID uuid.UUID, subscriptionID uuid.UUID, autoRenew bool) error {
	subscription, err := Get(userID, subscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return errors.New("subscription not found")
	}

	if subscription.StartAt.AddDate(0, 0, subscription.Subscription.ValidForInDays).Before(time.Now()) {
		return errors.New("subscription already expired")
	}

	// ~ Update the subscription
	if err := database.DB.Model(&subscription).Update("auto_renew", autoRenew).Error; err != nil {
		return err
	}

	return nil
}

func UpdatePrice(userSubscriptionID uuid.UUID, totalPrice int) error {
	subscription, err := GetActive(userSubscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return errors.New("subscription not found")
	}

	if subscription.StartAt.AddDate(0, 0, subscription.Subscription.ValidForInDays).Before(time.Now()) {
		return errors.New("subscription already expired")
	}

	// ~ Update the subscription
	if err := database.DB.Model(&subscription).Update("total_price", totalPrice).Error; err != nil {
		return err
	}

	return nil
}

// Cancelling a subscription is not deleting it, but making it innaccessible
// This is an admin action only.
// The user can't cancel/refund a subscription, but can only deactivate the auto-renewal
func Cancel(userID uuid.UUID, subscriptionID uuid.UUID) error {
	subscription, err := Get(userID, subscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return errors.New("subscription not found")
	}

	if subscription.StartAt.AddDate(0, 0, subscription.Subscription.ValidForInDays).Before(time.Now()) {
		return errors.New("subscription already expired")
	}

	// ~ Switch subscription to innaccessible
	if err := database.DB.Model(&subscription).Update("is_accessible", false).Error; err != nil {
		return err
	}

	return nil
}

// This function calculates the total price of a subscription, including perks
// In case of an error, it returns a very high price to avoid any issues
func CalculateTotalPrice(userSubscriptionID uuid.UUID) (int, error) {
	var total = 0
	// ~ Find the subscription
	var subscription models.UserSubscription
	if err := database.DB.Where("id = ?", userSubscriptionID).First(&subscription).Error; err != nil {
		return 100e10, err
	}

	total += subscription.Subscription.Price

	// ~ Find the subscription perks
	var subscriptionPerks models.SubscriptionPerks
	if err := database.DB.Where("user_subscription_id = ?", userSubscriptionID).First(&subscriptionPerks).Error; err != nil {
		return 100e10, err
	}

	if subscriptionPerks.CollaborativeTeamCount-subscriptionPerks.IncludedTeamCount > 0 {
		total += subscriptionPerks.PricePerAdditionalTeam * (subscriptionPerks.CollaborativeTeamCount - subscriptionPerks.IncludedTeamCount)
	}
	if subscriptionPerks.MaxProductsPerTeam-subscriptionPerks.IncludedProductCount > 0 {
		total += subscriptionPerks.PricePerAdditionalProduct * (subscriptionPerks.MaxProductsPerTeam - subscriptionPerks.IncludedProductCount)
	}

	return total, nil
}
