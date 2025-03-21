package admin_subscriptions

import (
	"encoding/json"
	"fmt"
	admin_subscription_service "gox/services/administration/subscriptions"
	subscriptions_service "gox/services/subscriptions"
	"gox/utils"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /administrate/subscriptions ~

func HandleGetSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := admin_subscription_service.GetAll()
	if err != nil {
		utils.AbortRequest(w, "Error fetching subscriptions", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, subs)
}

func HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Price          int    `json:"price"`
		Currency       string `json:"currency"`
		ValidForInDays int    `json:"valid_for_in_days"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, fmt.Sprintf("invalid body: %v", err), http.StatusBadRequest)
		return
	}

	sub, err := admin_subscription_service.Create(input.Name, input.Description, input.Price, input.Currency, input.ValidForInDays)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, sub)
}

// ~ /administrate/subscriptions/{id} ~

func getSubscriptionID(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	subID := vars["id"]
	subUUID, err := uuid.Parse(subID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("id invalid")
	}

	return subUUID, nil
}

func HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := getSubscriptionID(r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := subscriptions_service.GetByID(id)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, sub)
}

func HandleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := getSubscriptionID(r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Price          int    `json:"price"`
		Currency       string `json:"currency"`
		ValidForInDays int    `json:"valid_for_in_days"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	sub, err := subscriptions_service.GetByID(id)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub.Name = input.Name
	sub.Description = input.Description
	sub.Price = input.Price
	sub.Currency = input.Currency
	sub.ValidForInDays = input.ValidForInDays

	sub, err = admin_subscription_service.Update(sub)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, sub)
}

func HandleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := getSubscriptionID(r)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := admin_subscription_service.DeleteByID(id.String()); err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, "deleted")
}
