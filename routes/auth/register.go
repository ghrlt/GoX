package auth

import (
	"encoding/json"
	"fmt"
	user_service "gox/services/users"
	"gox/utils"
	"net/http"
)

// handleRegister crée un utilisateur, sa personal team, son team member et retourne un token JWT
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	utils.ConsoleLog("📥 Requête POST /auth/register")

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Créer l'utilisateur
	userID, err := user_service.Create(input.Email, input.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create user: %s", err), http.StatusInternalServerError)
		return
	}

	// Générer un token JWT
	token, err := utils.GenerateJWT(userID, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not generate token: %s", err), http.StatusInternalServerError)
		return
	}

	// Réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token": token,
	})

}
