package teams

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"gox/database/models"
	team_member_service "gox/services/teams/members"
	user_service "gox/services/users"
	"gox/utils"
)

// ~ /teams/{id}/members ~
func HandleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	var input struct {
		UserID string                `json:"user_id"`
		Role   models.TeamMemberRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		utils.AbortRequest(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Vérification de l'existence de l'utilisateur
	_, err = user_service.Get(userUUID)
	if err != nil {
		utils.AbortRequest(w, "User not found", http.StatusNotFound)
		return
	}

	// Ajout du membre à la Team
	err = team_member_service.Add(teamUUID, userUUID, input.Role)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

func HandleGetTeamMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	// Récupération des membres de la Team
	members, err := team_member_service.GetAll(teamUUID)
	if err != nil {
		utils.AbortRequest(w, "Error fetching team members", http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	data := make([]map[string]interface{}, len(members))
	for i, member := range members {
		data[i] = map[string]interface{}{
			"id":        member.ID,
			"user_id":   member.MemberID,
			"role":      member.Role,
			"is_active": member.IsActive,
		}
	}
	utils.RespondJSON(w, data)
}

// ~ /teams/{id}/members/{member_id} ~
func checkForMemberID(memberID string) (uuid.UUID, error) {
	if memberID == "" {
		return uuid.Nil, nil
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return uuid.Nil, err
	}

	return memberUUID, nil
}

func HandleGetTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]
	memberID := vars["member_id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		utils.AbortRequest(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Récupération du membre de la Team
	member, err := team_member_service.Get(teamUUID, memberID)
	if err != nil {
		utils.AbortRequest(w, "Team member not found", http.StatusNotFound)
		return
	}

	// Réponse JSON
	data := map[string]interface{}{
		"id":        member.ID,
		"user_id":   member.MemberID,
		"role":      member.Role,
		"is_active": member.IsActive,
	}
	utils.RespondJSON(w, data)
}

func HandleUpdateTeamMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]
	memberID := vars["member_id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	memberUUID, err := checkForMemberID(memberID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Invalid member ID: %v", err), http.StatusBadRequest)
		return
	}

	var input struct {
		Role models.TeamMemberRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.AbortRequest(w, "body invalid", http.StatusBadRequest)
		return
	}

	// Mise à jour du rôle du membre
	err = team_member_service.UpdateRole(teamUUID, memberUUID, input.Role)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

func HandleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]
	memberID := vars["member_id"]

	teamUUID, err := checkForTeamID(teamID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Invalid team ID: %v", err), http.StatusBadRequest)
		return
	}

	memberUUID, err := checkForMemberID(memberID)
	if err != nil {
		utils.AbortRequest(w, fmt.Sprintf("Invalid member ID: %v", err), http.StatusBadRequest)
		return
	}

	// Suppression du membre de la Team
	err = team_member_service.Remove(teamUUID, memberUUID)
	if err != nil {
		utils.AbortRequest(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
