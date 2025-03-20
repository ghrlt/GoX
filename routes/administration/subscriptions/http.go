package admin_subscriptions

import (
	"encoding/json"
	"fmt"
	admin_subscription_service "gox/services/administration"
	subscriptions_service "gox/services/subscriptions"
	"gox/utils"
	"net/http"

	"github.com/google/uuid"
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

func HandleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	input := struct {
		ID     uuid.UUID `json:"id"`
		Update struct {
			Name           string `json:"name"`
			Description    string `json:"description"`
			Price          int    `json:"price"`
			Currency       string `json:"currency"`
			ValidForInDays int    `json:"valid_for_in_days"`
		} `json:"update"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	sub, err := subscriptions_service.GetByID(input.ID)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub.Name = input.Update.Name
	sub.Description = input.Update.Description
	sub.Price = input.Update.Price
	sub.Currency = input.Update.Currency
	sub.ValidForInDays = input.Update.ValidForInDays

	sub, err = admin_subscription_service.Update(sub)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, sub)
}
