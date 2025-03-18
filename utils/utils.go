package utils

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		// ConsoleLog("🔑 %s=%s", key, value)
		return value
	}
	return fallback
}

type Logger struct {
	fatal bool
	err   bool
}

func ConsoleLog(msg string, args ...any) Logger {
	fmt.Println(fmt.Sprintf(msg, args...))
	return Logger{fatal: false, err: false}
}

func (l Logger) Fatal() {
	l.fatal = true
	if l.fatal {
		os.Exit(1)
	}
}
func (l Logger) Error() {
	l.err = true
	if l.err {

	}
}

const defaultJWTSecret = "unsecure-hard-coded-secret"

func GenerateJWT(userID uuid.UUID, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"user":  userID.String(),
		"admin": isAdmin,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Expire en 24h
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetEnv("JWT_SECRET", defaultJWTSecret)))
}

func DecodeJWT(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetEnv("JWT_SECRET", defaultJWTSecret)), nil
	})
	return claims, err
}

func ExtractUserIDFromJWT(r *http.Request) (uuid.UUID, error) {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		return uuid.UUID{}, fmt.Errorf("token is missing")
	}

	claims, err := DecodeJWT(tokenString)
	if err != nil {
		return uuid.UUID{}, err
	}

	userIDStr, ok := claims["user"].(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("user not found in token")
	}

	// Convertir string en uuid.UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID format: %v", err)
	}

	return userID, nil
}
