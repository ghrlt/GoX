package database

import (
	"fmt"

	"gox/database/models"
	"gox/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initialise la connexion à PostgreSQL
func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.ConsoleLog("❌ Erreur de connexion à la base de données : %v", err).Fatal()
	}

	// Migrations automatiques
	err = DB.AutoMigrate(
		&models.Team{},
		&models.TeamMember{},
		&models.User{},
		&models.UserProfile{},
		&models.UserCredit{},
		&models.UserSubscription{},
		&models.Subscription{},
		&models.Coupon{},
		&models.CouponHistory{},
		&models.RequestLog{},
	)
	if err != nil {
		utils.ConsoleLog("❌ Erreur lors des migrations : %v", err).Fatal()
	}

	fmt.Println("🚀 Connexion à la base de données établie")
}
