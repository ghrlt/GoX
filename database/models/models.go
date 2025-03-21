package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamType string

const (
	TeamTypePersonal TeamType = "personal"
	TeamTypeCompany  TeamType = "company"
)

type TeamMemberRole string

const (
	TeamMemberRoleOwner     TeamMemberRole = "owner"
	TeamMemberRoleAdmin     TeamMemberRole = "admin"
	TeamMemberRoleSpectator TeamMemberRole = "spectator"
)

type CreditOperationType string

const (
	CreditOperationTypeAdd    CreditOperationType = "add"
	CreditOperationTypeRemove CreditOperationType = "remove"
	CreditOperationTypeUse    CreditOperationType = "use"
)

type CouponTypes string

const (
	CouponTypesCredits CouponTypes = "credits"
)

type Team struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Type         TeamType  `gorm:"not null"`
	Name         string    `gorm:"not null"`
	IsAccessible bool      `gorm:"default:true"`
}
type TeamMember struct {
	ID           uint           `gorm:"primaryKey;autoIncrement"`
	MemberID     uuid.UUID      `gorm:"index;not null"`
	Member       User           `gorm:"foreignKey:MemberID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	TeamID       uuid.UUID      `gorm:"index;not null"`
	Team         Team           `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Role         TeamMemberRole `gorm:"not null"`
	IsActive     bool           `gorm:"default:true"`
	IsAccessible bool           `gorm:"default:true"`
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"index;unique"`
	Password     string    `gorm:"not null"`
	CreatedOn    time.Time `gorm:"autoCreateTime"`
	IsActive     bool      `gorm:"default:true"`
	IsAccessible bool      `gorm:"default:true"`
}

type UserProfile struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CustomerID uuid.UUID `gorm:"index;not null"`
	Customer   User      `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Username   string    `gorm:"not null"`
	// AvatarURL          string   `gorm:"default:null"`
	PublicStatsDisplay bool `gorm:"default:true"`
	IsAccessible       bool `gorm:"default:true"`
}

type UserCredit struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CustomerID   uuid.UUID `gorm:"index;not null"`
	Customer     User      `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	Balance      int       `gorm:"not null;default:0"`
	IsAccessible bool      `gorm:"default:true"`
}

type UserCreditHistory struct {
	ID           uint                `gorm:"primaryKey;autoIncrement"`
	CustomerID   uuid.UUID           `gorm:"index;not null"`
	Customer     User                `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	Amount       int                 `gorm:"not null"`
	Operation    CreditOperationType `gorm:"not null"`
	Reason       string              `gorm:"not null"`
	DateTime     time.Time           `gorm:"not null"`
	IsAccessible bool                `gorm:"default:true"`
}

type UserSubscription struct {
	ID                uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CustomerID        uuid.UUID         `gorm:"index;not null"`
	Customer          User              `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	SubscriptionID    uuid.UUID         `gorm:"index;not null"`
	Subscription      Subscription      `gorm:"foreignKey:SubscriptionID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	SubscriptionPerks SubscriptionPerks `gorm:"foreignKey:UserSubscriptionID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	StartAt           time.Time         `gorm:"not null"`
	AutoRenew         bool              `gorm:"default:true"`
	TotalPrice        int               `gorm:"not null"`
	IsAccessible      bool              `gorm:"default:true"`
}

type Subscription struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name           string    `gorm:"not null"`
	Description    string    `gorm:"not null"`
	Price          int       `gorm:"not null"`
	Currency       string    `gorm:"default:'credits'"`
	ValidForInDays int       `gorm:"default:7"`
	IsAccessible   bool      `gorm:"default:true"`
}

type SubscriptionPerks struct {
	ID                        uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserSubscriptionID        uuid.UUID         `gorm:"index;not null"`
	UserSubscription          *UserSubscription `gorm:"foreignKey:UserSubscriptionID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	CollaborativeTeamCount    int               `gorm:"default:1"`
	IncludedTeamCount         int               `gorm:"default:1"`
	PricePerAdditionalTeam    int               `gorm:"default:25"`
	MaxProductsPerTeam        int               `gorm:"default:1"`
	IncludedProductCount      int               `gorm:"default:1"`
	PricePerAdditionalProduct int               `gorm:"default:50"`
	IsAccessible              bool              `gorm:"default:true"`
}

type RequestLog struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	UserID    *uuid.UUID `gorm:"index;default:null"`
	Domain    string     `gorm:"index"`
	Endpoint  string     `gorm:"index"`
	Content   string     `gorm:"type:bytea"`
	Method    string
	Status    int
	Timestamp time.Time `gorm:"autoCreateTime"`
}

func (r *RequestLog) BeforeCreate(tx *gorm.DB) (err error) {
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	return
}
