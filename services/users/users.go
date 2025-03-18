package user_service

import (
	"fmt"
	"gox/database"
	"gox/database/models"
	team_service "gox/services/teams"
	team_member_service "gox/services/teams/members"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Create(email, password string) (uuid.UUID, error) {
	// Vérification des champs requis
	if email == "" || password == "" {
		return uuid.UUID{}, fmt.Errorf("email and password are required")
	}
	// Vérification de l'unicité de l'email
	var count int64
	if err := database.DB.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return uuid.UUID{}, fmt.Errorf("error checking email: %v", err)
	}
	if count > 0 {
		return uuid.UUID{}, fmt.Errorf("email already used")
	}

	// Hash du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error hashing password: %v", err)
	}

	// Création de l'utilisateur
	user := models.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	// Insertion en base
	if err := database.DB.Create(&user).Error; err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating user: %v", err)
	}

	// Création d'une équipe par défaut pour l'utilisateur
	teamID, err := team_service.Create("Personal", models.TeamTypePersonal)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating default team: %v", err)
	}

	// Création d'un team member pour l'utilisateur dans l'équipe par défaut
	err = team_member_service.Add(teamID, user.ID, models.TeamMemberRoleOwner)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating default team member: %v", err)
	}

	// retourner l'utilisateur créé
	return user.ID, nil
}

func Get(userID uuid.UUID) (models.User, error) {
	// Récupération de l'utilisateur
	var user models.User
	result := database.DB.Where("id = ?", userID).First(&user)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return models.User{}, result.Error
	}

	return user, nil
}

func GetByEmail(email string) (models.User, error) {
	// Récupération de l'utilisateur
	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return models.User{}, result.Error
	}

	return user, nil
}

func UpdateEmail(userID uuid.UUID, email string) error {
	// Vérification des champs requis
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Vérification de l'unicité de l'email
	var count int64
	if err := database.DB.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return fmt.Errorf("error checking email: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("email already used")
	}

	// Mise à jour de l'utilisateur
	result := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("email", email)

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdatePassword(userID uuid.UUID, password string) error {
	// Vérification des champs requis
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Hash du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	// Mise à jour de l'utilisateur
	result := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", string(hashedPassword))

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func Delete(userID uuid.UUID) error {
	// Suppression de l'utilisateur
	result := database.DB.Where("id = ?", userID).Delete(&models.User{})

	// Vérification des erreurs GORM
	if result.Error != nil {
		return result.Error
	}

	return nil
}
