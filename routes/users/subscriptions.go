package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gox/database/models"
	user_subscription_service "gox/services/users/subscriptions"
	user_sub_perks_service "gox/services/users/subscriptions/perks"
	"gox/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /users/{id}/subscriptions ~
func HandleGetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Détermination du type d'abonnement demandé
	var userSubscriptions []models.UserSubscription
	queryParam := r.URL.Query().Get("current")

	switch queryParam {
	case "true":
		// Récupération de l'abonnement actuel
		userSubscription, err := user_subscription_service.GetActive(userUUID)
		if err != nil {
			utils.AbortRequest(w, "Error fetching user current subscription", http.StatusInternalServerError)
			return
		}
		userSubscriptions = []models.UserSubscription{*userSubscription}

	case "false":
		// Récupération des anciens abonnements
		userSubscriptions, err = user_subscription_service.GetInactives(userUUID)
		if err != nil {
			utils.AbortRequest(w, "Error fetching user old subscriptions", http.StatusInternalServerError)
			return
		}

	default:
		// Récupération de tous les abonnements
		userSubscriptions, err = user_subscription_service.GetAll(userUUID)
		if err != nil {
			utils.AbortRequest(w, "Error fetching user subscriptions", http.StatusInternalServerError)
			return
		}
	}

	// Struct de réponse unique
	type SubscriptionResponse struct {
		UserSubscriptionID uuid.UUID `json:"user_subscription_id"`
		SubscriptionID     uuid.UUID `json:"subscription_id"`
		CustomerID         uuid.UUID `json:"customer_id"`
		TotalPrice         int       `json:"total_price"`
		AutoRenew          bool      `json:"auto_renew"`
		StartAt            string    `json:"start_at"`
		Perks              struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		} `json:"perks"`
	}

	// Transformation des abonnements en format JSON
	responseData := make([]SubscriptionResponse, len(userSubscriptions))
	for i, userSubscription := range userSubscriptions {
		responseData[i] = SubscriptionResponse{
			UserSubscriptionID: userSubscription.ID,
			SubscriptionID:     userSubscription.SubscriptionID,
			CustomerID:         userSubscription.CustomerID,
			TotalPrice:         userSubscription.TotalPrice,
			AutoRenew:          userSubscription.AutoRenew,
			StartAt:            userSubscription.StartAt.String(),
			Perks: struct {
				CollaborativeTeamCount int `json:"collaborative_team_count"`
				MaxProductsPerTeam     int `json:"max_products_per_team"`
			}{
				CollaborativeTeamCount: userSubscription.SubscriptionPerks.CollaborativeTeamCount,
				MaxProductsPerTeam:     userSubscription.SubscriptionPerks.MaxProductsPerTeam,
			},
		}
	}

	// Réponse JSON
	utils.RespondJSON(w, responseData)
}

func HandleCreateUserSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Vérification de la non-existence d'un abonnement actif
	activeSubscription, err := user_subscription_service.GetActive(userUUID)
	if err != nil {
		utils.AbortRequest(w, "Error fetching user active subscription", http.StatusInternalServerError)
		return
	}
	if activeSubscription != nil {
		utils.AbortRequest(w, "User already has an active subscription", http.StatusBadRequest)
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
		utils.AbortRequest(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Création de l'abonnement
	userSubscription, err := user_subscription_service.Create(userUUID, newUserSubscriptionData.SubscriptionID, newUserSubscriptionData.AutoRenew)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Error creating user subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Ajout des avantages de l'abonnement
	err = user_sub_perks_service.AddPerks(userUUID, userSubscription.ID, newUserSubscriptionData.Perks.CollaborativeTeamCount, newUserSubscriptionData.Perks.MaxProductsPerTeam)
	if err != nil {
		utils.AbortRequest(w, "Error adding subscription perks", http.StatusInternalServerError)
		return
	}

	// Calcul du prix total
	totalPrice, err := user_subscription_service.CalculateTotalPrice(userSubscription.ID)
	if err != nil {
		utils.AbortRequest(w, "Error calculating total price", http.StatusInternalServerError)
		return
	}

	// Mise à jour du prix total
	if err := user_subscription_service.UpdatePrice(userSubscription.ID, totalPrice); err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Error updating user subscription price: %v", err), http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	data := struct {
		UserSubscriptionID uuid.UUID `json:"user_subscription_id"`
		SubscriptionID     uuid.UUID `json:"subscription_id"`
		CustomerID         uuid.UUID `json:"customer_id"`
		TotalPrice         int       `json:"total_price"`
		AutoRenew          bool      `json:"auto_renew"`
		StartAt            string    `json:"start_at"`
		Perks              struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		} `json:"perks"`
	}{
		UserSubscriptionID: userSubscription.ID,
		SubscriptionID:     userSubscription.SubscriptionID,
		CustomerID:         userSubscription.CustomerID,
		TotalPrice:         userSubscription.TotalPrice,
		AutoRenew:          userSubscription.AutoRenew,
		StartAt:            userSubscription.StartAt.String(),
		Perks: struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		}{
			CollaborativeTeamCount: newUserSubscriptionData.Perks.CollaborativeTeamCount,
			MaxProductsPerTeam:     newUserSubscriptionData.Perks.MaxProductsPerTeam,
		},
	}
	utils.RespondJSON(w, data)
}

// ~ /users/{id}/subscriptions/{subscription_id} ~
func getSubscriptionID(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	subscriptionID := vars["subscription_id"]
	if subscriptionID == "" {
		utils.AbortRequest(w, "subscription id is required", http.StatusBadRequest)
		return uuid.Nil, fmt.Errorf("subscription id is required")
	}

	subscriptionUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		utils.AbortRequest(w, "invalid subscription id", http.StatusBadRequest)
		return uuid.Nil, fmt.Errorf("invalid subscription id")
	}

	return subscriptionUUID, nil
}

func HandleGetUserSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	subscriptionUUID, err := getSubscriptionID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	// Récupération de l'abonnement
	userSubscription, err := user_subscription_service.Get(userUUID, subscriptionUUID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Error fetching user subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	data := struct {
		UserSubscriptionID uuid.UUID `json:"user_subscription_id"`
		SubscriptionID     uuid.UUID `json:"subscription_id"`
		CustomerID         uuid.UUID `json:"customer_id"`
		TotalPrice         int       `json:"total_price"`
		AutoRenew          bool      `json:"auto_renew"`
		StartAt            string    `json:"start_at"`
		ValidUntil         string    `json:"valid_until"`
		Perks              struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		} `json:"perks"`
	}{
		UserSubscriptionID: userSubscription.ID,
		SubscriptionID:     userSubscription.SubscriptionID,
		CustomerID:         userSubscription.CustomerID,
		TotalPrice:         userSubscription.TotalPrice,
		AutoRenew:          userSubscription.AutoRenew,
		StartAt:            userSubscription.StartAt.String(),
		ValidUntil:         userSubscription.StartAt.AddDate(0, 0, userSubscription.Subscription.ValidForInDays).String(),
		Perks: struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		}{
			CollaborativeTeamCount: userSubscription.SubscriptionPerks.CollaborativeTeamCount,
			MaxProductsPerTeam:     userSubscription.SubscriptionPerks.MaxProductsPerTeam,
		},
	}
	utils.RespondJSON(w, data)
}

func HandleUpdateUserSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	subscriptionUUID, err := getSubscriptionID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	// Obtention des données de l'abonnement
	var updateSubscriptionData struct {
		AutoRenew bool `json:"auto_renew"`
	}
	err = json.NewDecoder(r.Body).Decode(&updateSubscriptionData)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Mise à jour de l'abonnement
	err = user_subscription_service.Update(userUUID, subscriptionUUID, updateSubscriptionData.AutoRenew)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Error updating user subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse, avec l'abonnement mis à jour
	userSubscription, err := user_subscription_service.Get(userUUID, subscriptionUUID)
	if err != nil {
		utils.AbortRequest(w, "Error fetching user subscription", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserSubscriptionID uuid.UUID `json:"user_subscription_id"`
		SubscriptionID     uuid.UUID `json:"subscription_id"`
		CustomerID         uuid.UUID `json:"customer_id"`
		TotalPrice         int       `json:"total_price"`
		AutoRenew          bool      `json:"auto_renew"`
		StartAt            string    `json:"start_at"`
		ValidUntil         string    `json:"valid_until"`
		Perks              struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		} `json:"perks"`
	}{
		UserSubscriptionID: userSubscription.ID,
		SubscriptionID:     userSubscription.SubscriptionID,
		CustomerID:         userSubscription.CustomerID,
		TotalPrice:         userSubscription.TotalPrice,
		AutoRenew:          userSubscription.AutoRenew,
		StartAt:            userSubscription.StartAt.String(),
		ValidUntil:         userSubscription.StartAt.AddDate(0, 0, userSubscription.Subscription.ValidForInDays).String(),
		Perks: struct {
			CollaborativeTeamCount int `json:"collaborative_team_count"`
			MaxProductsPerTeam     int `json:"max_products_per_team"`
		}{
			CollaborativeTeamCount: userSubscription.SubscriptionPerks.CollaborativeTeamCount,
			MaxProductsPerTeam:     userSubscription.SubscriptionPerks.MaxProductsPerTeam,
		},
	}
	utils.RespondJSON(w, data)

}

func HandleDeleteUserSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.AbortRequest(w, "invalid user id", http.StatusBadRequest)
		return
	}

	subscriptionUUID, err := getSubscriptionID(w, r)
	if err != nil {
		utils.AbortRequest(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	// Suppression de l'abonnement
	err = user_subscription_service.Cancel(userUUID, subscriptionUUID)
	if err != nil {
		utils.AbortRequest(w, "Error cancelling user subscription", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	utils.RespondJSON(w, "User subscription cancelled")
}
