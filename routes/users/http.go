package users

import (
	"encoding/json"
	"errors"
	"gox/database"
	"gox/database/models"
	auth_utils "gox/services/auth"
	team_service "gox/services/teams"
	user_service "gox/services/users"
	user_profile_service "gox/services/users/profile"
	"gox/utils"
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
		utils.AbortRequest(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Création de l'utilisateur
	user, err := user_service.Create(input.Email, input.Password)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	utils.RespondJSON(w, map[string]interface{}{
		"success": true,
		"user_id": user.ID,
	})
}

func HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		utils.AbortRequest(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	data := make([]map[string]interface{}, len(users))
	for i, user := range users {
		data[i] = map[string]interface{}{
			"id":            user.ID.String(),
			"email":         user.Email,
			"created_on":    user.CreatedOn,
			"is_active":     user.IsActive,
			"is_accessible": user.IsAccessible,
		}
	}
}

// ~ /users/{id} ~
func getUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		return "", errors.New("user id is required")
	}

	if userID == "me" {
		// Récupérer l'utilisateur actuel, à partir du token JWT
		userID = auth_utils.GetAuthenticatedUserID(w, r)
		if userID == "" {
			return "", errors.New("unauthorized")
		}
	}

	return userID, nil
}

func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusNotFound)
		return
	}

	// Réponse JSON
	data := struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		IsActive bool   `json:"is_active"`
	}{
		ID:       user.ID.String(),
		Email:    user.Email,
		IsActive: user.IsActive,
	}
	utils.RespondJSON(w, data)
}

func HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusNotFound)
		return
	}

	// Lire le corps de la requête, pour obtenir les données à mettre à jour (email ou password)
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Mettre à jour l'utilisateur
	if Email := input.Email; Email != "" {
		if err := user_service.UpdateEmail(user.ID, Email); err != nil {
			utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if Password := input.Password; Password != "" {
		if err := user_service.UpdatePassword(user.ID, Password); err != nil {
			utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Réponse JSON
	utils.RespondJSON(w, map[string]interface{}{
		"success": true,
	})
}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Supprimer l'utilisateur
	if err := user_service.Delete(userUUID); err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	utils.RespondJSON(w, map[string]interface{}{
		"success": true,
	})
}

// ~ /users/{id}/teams ~

func HandleGetUserTeams(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Récupérer les équipes de l'utilisateur
	teams, err := team_service.GetTeamsByMemberID(userUUID)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(teams) == 0 {
		utils.AbortRequest(w, "No teams found", http.StatusNotFound)
		return
	}

	// Réponse JSON
	data := make([]map[string]interface{}, len(teams))
	for i, team := range teams {
		data[i] = map[string]interface{}{
			"id":   team.ID.String(),
			"name": team.Name,
			"type": team.Type,
		}
	}
	utils.RespondJSON(w, data)
}

// ~ /users/{id}/profile ~

func HandleCreateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Lire le corps de la requête, pour obtenir les données du profil
	var profileData struct {
		Username string `json:"username"`
		// AvatarURL          string `json:"avatar_url"`
		PublicStatsDisplay bool `json:"public_stats_display"`
	}
	if err := json.NewDecoder(r.Body).Decode(&profileData); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	userProfile := models.UserProfile{
		Username: profileData.Username,
		// AvatarURL:          profileData.AvatarURL,
		PublicStatsDisplay: profileData.PublicStatsDisplay,
	}

	// Création du profil de l'utilisateur
	if err := user_profile_service.Create(userUUID, userProfile); err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	utils.RespondJSON(w, map[string]interface{}{
		"success":    true,
		"profile_id": userProfile.ID,
	})
}

func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Vérifier que l'user ID existe
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	_, err = user_service.Get(userUUID)
	if err != nil {
		utils.AbortRequest(w, "User not found", http.StatusNotFound)
		return
	}

	// Récupérer le profil de l'utilisateur
	var user models.UserProfile
	if err := database.DB.Where("customer_id = ?", userID).First(&user).Error; err != nil {
		utils.AbortRequest(w, "User profile not found", http.StatusNotFound)
		return
	}

	// Vérifier si le profil est accessible
	if !user.IsAccessible {
		utils.AbortRequest(w, "User profile is not accessible", http.StatusForbidden)
		return
	}

	// Réponse JSON
	data := struct {
		CustomerID string `json:"customer_id"`
		Username   string `json:"username"`
		// AvatarURL          string `json:"avatar_url"`
		PublicStatsDisplay bool `json:"public_stats_display"`
	}{
		CustomerID: user.CustomerID.String(),
		Username:   user.Username,
		// AvatarURL:          user.AvatarURL,
		PublicStatsDisplay: user.PublicStatsDisplay,
	}
	utils.RespondJSON(w, data)
}

func HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le profil de l'utilisateur
	var user models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		utils.AbortRequest(w, "User profile not found", http.StatusNotFound)
		return
	}

	var input struct {
		Username string `json:"username"`
		// AvatarURL          string `json:"avatar_url"`
		PublicStatsDisplay bool `json:"public_stats_display"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Mettre à jour le profil de l'utilisateur
	profile := models.UserProfile{
		Username: input.Username,
		// AvatarURL:          input.AvatarURL,
		PublicStatsDisplay: input.PublicStatsDisplay,
	}
	if err := user_profile_service.Update(user.ID, profile); err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	utils.RespondJSON(w, map[string]interface{}{
		"success": true,
	})
}

func HandleDeleteUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// Supprimer le profil de l'utilisateur
	if err := user_profile_service.Delete(userUUID); err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	utils.RespondJSON(w, map[string]interface{}{
		"success": true,
	})
}
