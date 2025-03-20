package auth_utils

import (
	"gox/utils"
	"net/http"
)

func CheckAuthenticationHeader(w http.ResponseWriter, r *http.Request) bool {
	// Récupérer le token
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.AbortRequest(w, "Authorization Token is missing.", http.StatusUnauthorized)
		return false
	}

	// Décoder le token et récupérer les claims
	claims, err := utils.DecodeJWT(tokenString)
	if err != nil {
		utils.AbortRequest(w, "Authorization Token is invalid.", http.StatusUnauthorized)
		return false
	}

	// Récupérer l'ID utilisateur
	authUserID, ok := claims["user"].(string)
	if !ok || authUserID == "" {
		utils.AbortRequest(w, "Authorization Token payload is invalid.", http.StatusUnauthorized)
		return false
	}

	// Log de l'utilisateur authentifié
	utils.ConsoleLog("🔑 Utilisateur authentifié: %s -> %s %s", authUserID, r.Method, r.URL.Path)

	return true
}

func IsAuthenticatedUserAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !CheckAuthenticationHeader(w, r) {
		return false
	}

	// Vérifier le rôle
	claims, err := utils.DecodeJWT(r.Header.Get("Authorization"))
	if err != nil {
		utils.AbortRequest(w, "Authorization Token is invalid.", http.StatusUnauthorized)
		return false
	}

	admin, ok := claims["admin"].(bool)
	if !ok || !admin {
		// utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
		return false
	}

	return true
}

func GetAuthenticatedUserID(w http.ResponseWriter, r *http.Request) string {
	// Récupérer le token
	tokenString := r.Header.Get("Authorization")

	// Décoder le token et récupérer les claims
	claims, err := utils.DecodeJWT(tokenString)
	if err != nil {
		return ""
	}

	// Récupérer l'ID utilisateur
	authUserID, ok := claims["user"].(string)
	if !ok || authUserID == "" {
		return ""
	}

	return authUserID
}
