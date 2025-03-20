package users

import (
	"encoding/json"
	user_subscription_service "gox/services/users/subscriptions"
	user_sub_perks_service "gox/services/users/subscriptions/perks"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /users/{id}/subscriptions/{subscription_id}/perks ~

func HandleGetUserSubscriptionPerks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	subscriptionID := vars["subscription_id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	if subscriptionID == "" {
		http.Error(w, "subscription id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	subscriptionUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	userSubscriptionPerks, err := user_sub_perks_service.GetPerks(userUUID, subscriptionUUID)
	if err != nil {
		http.Error(w, "Error fetching user subscription perks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userSubscriptionPerks)
}

func HandleUpdateUserSubscriptionPerks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	subscriptionID := vars["subscription_id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	if subscriptionID == "" {
		http.Error(w, "subscription id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	subscriptionUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	var updateSubscriptionData struct {
		CollaborativeTeamCount int `json:"collaborative_team_count"`
		MaxProductsPerTeam     int `json:"max_products_per_team"`
	}

	// Mise à jour des avantages de l'abonnement
	err = user_sub_perks_service.UpdatePerks(userUUID, subscriptionUUID, updateSubscriptionData.CollaborativeTeamCount, updateSubscriptionData.MaxProductsPerTeam)
	if err != nil {
		http.Error(w, "Error updating subscription perks", http.StatusInternalServerError)
		return
	}

	// Mise à jour du prix total
	totalPrice, err := user_subscription_service.CalculateTotalPrice(subscriptionUUID)
	if err != nil {
		http.Error(w, "Error calculating total price", http.StatusInternalServerError)
		return
	}

	if err := user_subscription_service.UpdatePrice(subscriptionUUID, totalPrice); err != nil {
		http.Error(w, "Error updating user subscription price", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse, avec les avantages mis à jour
	userSubscriptionPerks, err := user_sub_perks_service.GetPerks(userUUID, subscriptionUUID)
	if err != nil {
		http.Error(w, "Error fetching user subscription perks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userSubscriptionPerks)
}
