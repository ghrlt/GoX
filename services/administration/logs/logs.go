package admin_logs_service

import (
	"gox/database"
	"gox/database/models"
)

func GetAll() ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetByDomain(domain string) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Where("domain = ?", domain).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetByEndpoint(endpoint string) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Where("endpoint = ?", endpoint).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetByMethod(method string) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Where("method = ?", method).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetByStatus(status int) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Where("status = ?", status).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetByDateRange(start, end string) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Where("timestamp BETWEEN ? AND ?", start, end).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetByUserID(userID string) ([]models.RequestLog, error) {
	var logs []models.RequestLog
	if err := database.DB.Where("user_id = ?", userID).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
