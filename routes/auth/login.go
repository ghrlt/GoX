package auth

import (
	"encoding/json"
	"fmt"
	"gox/database"
	"gox/database/models"
	"gox/utils"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// handleLogin vérifie les credentials et retourne un token JWT
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Vérifie l’utilisateur en base
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Vérifie le mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Générer un token JWT
	token, err := utils.GenerateJWT(user.ID, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not generate token: %s", err), http.StatusInternalServerError)
		return
	}

	// Réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
	})
}
