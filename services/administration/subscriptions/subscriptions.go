package admin_subscription_service

import (
	"gox/database"
	"gox/database/models"
)

func GetAll() ([]models.Subscription, error) {
	var subs []models.Subscription
	err := database.DB.Find(&subs).Error
	return subs, err
}

func Create(name, description string, price int, currency string, validForInDays int) (models.Subscription, error) {
	sub := models.Subscription{
		Name:           name,
		Description:    description,
		Price:          price,
		Currency:       currency,
		ValidForInDays: validForInDays,
	}
	err := database.DB.Create(&sub).Error
	return sub, err
}

func Update(subscription models.Subscription) (models.Subscription, error) {
	err := database.DB.Save(&subscription).Error
	return subscription, err
}

func Delete(subscription models.Subscription) error {
	return database.DB.Delete(&subscription).Error
}

func DeleteByID(id string) error {
	return database.DB.Delete(&models.Subscription{}, id).Error
}
