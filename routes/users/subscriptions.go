package users

import (
	"encoding/json"
	"net/http"

	user_subscription_service "gox/services/users/subscriptions"
	user_sub_perks_service "gox/services/users/subscriptions/perks"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /users/{id}/subscriptions ~

func HandleGetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("current") == "true" {
		// Récupération de l'abonnement actuel de l'utilisateur
		userSubscription, err := user_subscription_service.GetActive(userUUID)
		if err != nil {
			http.Error(w, "Error fetching user current subscription", http.StatusInternalServerError)
			return
		}

		// Envoi de la réponse
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscription)
		return
	} else if r.URL.Query().Get("current") == "false" {
		// Récupération des anciens abonnements de l'utilisateur
		userSubscriptions, err := user_subscription_service.GetInactives(userUUID)
		if err != nil {
			http.Error(w, "Error fetching user old subscriptions", http.StatusInternalServerError)
			return
		}

		// Envoi de la réponse
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscriptions)
		return
	} else {
		// Récupération de tout les abonnements de l'utilisateur
		userSubscriptions, err := user_subscription_service.GetAll(userUUID)
		if err != nil {
			http.Error(w, "Error fetching user subscriptions", http.StatusInternalServerError)
			return
		}

		// Envoi de la réponse
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscriptions)
	}
}

func HandleCreateUserSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Obtention des données de l'abonnement
	var newUserSubscriptionData struct {
		SubscriptionID uuid.UUID `json:"subscription_id"`
		AutoRenew      bool      `json:"auto_renew"`
		Perks          struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		} `json:"perks"`
	}
	err = json.NewDecoder(r.Body).Decode(&newUserSubscriptionData)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Création de l'abonnement
	userSubscription, err := user_subscription_service.Create(userUUID, newUserSubscriptionData.SubscriptionID, newUserSubscriptionData.AutoRenew)
	if err != nil {
		http.Error(w, "Error creating user subscription", http.StatusInternalServerError)
		return
	}

	// Ajout des avantages de l'abonnement
	err = user_sub_perks_service.AddPerks(userUUID, userSubscription.ID, newUserSubscriptionData.Perks.CollaborativeTeamCount, newUserSubscriptionData.Perks.MaxProductsPerTeam)
	if err != nil {
		http.Error(w, "Error adding subscription perks", http.StatusInternalServerError)
		return
	}

	// Calcul du prix total
	totalPrice, err := user_subscription_service.CalculateTotalPrice(userSubscription.ID)
	if err != nil {
		http.Error(w, "Error calculating total price", http.StatusInternalServerError)
		return
	}

	// Mise à jour du prix total
	if err := user_subscription_service.UpdatePrice(userSubscription.ID, totalPrice); err != nil {
		http.Error(w, "Error updating user subscription price", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userSubscription)
}

// ~ /users/{id}/subscriptions/{subscription_id} ~

func HandleGetUserSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	subscriptionID := vars["subscription_id"]

	if userID == "" || subscriptionID == "" {
		http.Error(w, "user id and subscription id are required", http.StatusBadRequest)
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

	// Récupération de l'abonnement
	userSubscription, err := user_subscription_service.Get(userUUID, subscriptionUUID)
	if err != nil {
		http.Error(w, "Error fetching user subscription", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userSubscription)
}

func HandleUpdateUserSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	subscriptionID := vars["subscription_id"]

	if userID == "" || subscriptionID == "" {
		http.Error(w, "user id and subscription id are required", http.StatusBadRequest)
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

	// Obtention des données de l'abonnement
	var updateSubscriptionData struct {
		AutoRenew bool `json:"auto_renew"`
	}
	err = json.NewDecoder(r.Body).Decode(&updateSubscriptionData)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Mise à jour de l'abonnement
	err = user_subscription_service.Update(userUUID, subscriptionUUID, updateSubscriptionData.AutoRenew)
	if err != nil {
		http.Error(w, "Error updating user subscription", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse, avec l'abonnement mis à jour
	userSubscription, err := user_subscription_service.Get(userUUID, subscriptionUUID)
	if err != nil {
		http.Error(w, "Error fetching user subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userSubscription)
}

func HandleDeleteUserSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	subscriptionID := vars["subscription_id"]

	if userID == "" || subscriptionID == "" {
		http.Error(w, "user id and subscription id are required", http.StatusBadRequest)
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

	// Suppression de l'abonnement
	err = user_subscription_service.Cancel(userUUID, subscriptionUUID)
	if err != nil {
		http.Error(w, "Error cancelling user subscription", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user subscription cancelled"})
}
