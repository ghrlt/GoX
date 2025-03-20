package teams

import (
	"encoding/json"
	"fmt"
	"gox/database"
	"gox/database/models"
	team_service "gox/services/teams"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /teams ~
func HandleCreateTeam(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string          `json:"name"`
		Type models.TeamType `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Création de la Team
	team, err := team_service.Create(input.Name, input.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"team_id": team.ID,
	})
}

func HandleViewTeams(w http.ResponseWriter, r *http.Request) {
	// Récupération de tous les Teams
	teams, err := team_service.GetAll(database.DB)
	if err != nil {
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// ~ /teams/{id} ~
func checkForTeamID(teamID string) (uuid.UUID, error) {
	if teamID == "" {
		return uuid.Nil, nil
	}

	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		return uuid.Nil, err
	}

	return teamUUID, nil
}

func HandleGetTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	// Récupération de la Team
	team, err := team_service.Get(teamUUID)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func HandleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Mise à jour de la Team
	err = team_service.UpdateName(teamUUID, input.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

func HandleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	// Suppression de la Team
	err = team_service.Delete(teamUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
