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

type CouponTypes string

const (
	CouponTypesCredits CouponTypes = "credits"
)

type Team struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Type         TeamType
	Name         string
	IsAccessible bool `gorm:"default:true"`
}
type TeamMember struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	MemberID     uuid.UUID `gorm:"index;not null"`
	Member       User      `gorm:"foreignKey:MemberID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	TeamID       uuid.UUID `gorm:"index;not null"`
	Team         Team      `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Role         TeamMemberRole
	IsActive     bool
	IsAccessible bool `gorm:"default:true"`
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"index;unique"`
	Password     string
	IsActive     bool
	IsAccessible bool `gorm:"default:true"`
}

type UserProfile struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CustomerID         uuid.UUID `gorm:"index;not null"`
	Customer           User      `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Username           string
	AvatarURL          string
	PublicStatsDisplay bool
	IsAccessible       bool `gorm:"default:true"`
}

type UserCredit struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CustomerID   uuid.UUID `gorm:"index;not null"`
	Customer     User      `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	Balance      int
	IsAccessible bool `gorm:"default:true"`
}

type UserSubscription struct {
	ID             uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CustomerID     uuid.UUID    `gorm:"index;not null"`
	Customer       User         `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	SubscriptionID uuid.UUID    `gorm:"index;not null"`
	Subscription   Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	StartedOn      time.Time
	ExpiresOn      time.Time
	IsAccessible   bool `gorm:"default:true"`
}
type Subscription struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string
	Description  string
	Price        int
	Currency     string
	ValidFor     int
	IsAccessible bool `gorm:"default:true"`
}

type Coupon struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SecretCode   string    `gorm:"unique"`
	Libelle      string
	Description  string
	Type         CouponTypes
	Value        int
	IssuedAt     time.Time
	Expires      time.Time
	IsUsed       bool
	IsAccessible bool `gorm:"default:true"`
}
type CouponHistory struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CouponID     uuid.UUID `gorm:"index;not null"`
	Coupon       Coupon    `gorm:"foreignKey:CouponID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	CustomerID   uuid.UUID `gorm:"index;not null"`
	Customer     User      `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
	UsedAt       time.Time
	IsAccessible bool `gorm:"default:true"`
}

type RequestLog struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	UserID    *uuid.UUID `gorm:"index;default:null"`
	Domain    string     `gorm:"index"`
	Endpoint  string     `gorm:"index"`
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
