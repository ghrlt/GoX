package users

import (
	"encoding/json"
	"gox/database"
	"gox/database/models"
	user_service "gox/services/users"
	user_profile_service "gox/services/users/profile"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /users ~
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Création de l'utilisateur
	user, err := user_service.Create(input.Email, input.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user_id": user,
	})

}

func HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ~ /users/{id} ~

func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		http.Error(w, "Invalid user id", http.StatusNotFound)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		http.Error(w, "Invalid user id", http.StatusNotFound)
		return
	}

	// Lire le corps de la requête, pour obtenir les données à mettre à jour (email ou password)
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Mettre à jour l'utilisateur
	if Email := input.Email; Email != "" {
		if err := user_service.UpdateEmail(user.ID, Email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if Password := input.Password; Password != "" {
		if err := user_service.UpdatePassword(user.ID, Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Supprimer l'utilisateur
	if err := user_service.Delete(userUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// ~ /users/{id}/profile ~

func HandleCreateUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Lire le corps de la requête, pour obtenir les données du profil
	var profileData struct {
		Username           string `json:"username"`
		AvatarURL          string `json:"avatar_url"`
		PublicStatsDisplay bool   `json:"public_stats_display"`
	}
	if err := json.NewDecoder(r.Body).Decode(&profileData); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	userProfile := models.UserProfile{
		Username:           profileData.Username,
		AvatarURL:          profileData.AvatarURL,
		PublicStatsDisplay: profileData.PublicStatsDisplay,
	}

	// Création du profil de l'utilisateur
	if err := user_profile_service.Create(database.DB, userUUID, userProfile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	// Récupérer le profil de l'utilisateur
	var user models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		http.Error(w, "Invalid user id", http.StatusNotFound)
		return
	}

	// Vérifier si le profil est accessible
	if !user.IsAccessible {
		http.Error(w, "This profile is not accessible", http.StatusLocked)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	// Récupérer le profil de l'utilisateur
	var user models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		http.Error(w, "Invalid user id", http.StatusNotFound)
		return
	}

	var input struct {
		Username           string `json:"username"`
		AvatarURL          string `json:"avatar_url"`
		PublicStatsDisplay bool   `json:"public_stats_display"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Mettre à jour le profil de l'utilisateur
	profile := models.UserProfile{
		Username:           input.Username,
		AvatarURL:          input.AvatarURL,
		PublicStatsDisplay: input.PublicStatsDisplay,
	}
	if err := user_profile_service.Update(database.DB, user.ID, profile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func HandleDeleteUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Supprimer le profil de l'utilisateur
	if err := user_profile_service.Delete(database.DB, userUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
