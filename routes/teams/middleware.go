package teams

import (
	"fmt"
	"gox/database/models"
	auth_utils "gox/services/auth"
	team_service "gox/services/teams"
	team_member_service "gox/services/teams/members"
	"gox/utils"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ~ /teams ~
func TeamsRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ~ Let's check that the user is authenticated
		if !auth_utils.CheckAuthenticationHeader(w, r) {
			return
		}

		if r.Method == http.MethodGet {
			// ~ Only admins can view all teams
			if !auth_utils.IsAuthenticatedUserAdmin(w, r) {
				utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
				return
			}
		}

		// ~ OK. Serve.
		next.ServeHTTP(w, r)
	})
}

// ~ /teams/{id} ~
// ~ /teams/{id}/members ~
func getTeamUUIDFromRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	teamID := vars["id"]
	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		utils.AbortRequest(w, "Invalid team ID", http.StatusBadRequest)
		return uuid.UUID{}, err
	}
	return teamUUID, nil
}

func TeamRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ~ Let's check that the user is authenticated
		if !auth_utils.CheckAuthenticationHeader(w, r) {
			return
		}

		userUUID, err := utils.ExtractUserIDFromJWT(r)
		if err != nil {
			utils.AbortRequest(w, "Authorization Token is invalid.", http.StatusUnauthorized)
			return
		}

		teamUUID, err := getTeamUUIDFromRequest(w, r)
		if err != nil {
			return
		}

		// ~ Check if the user is a member of the team
		isIn, err := team_service.IsUserInTeam(userUUID, teamUUID)
		if err != nil {
			utils.AbortRequest(w, "An error occured", http.StatusInternalServerError)
			return
		}
		if !isIn {
			utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// ~ Check team member permissions
		if r.Method == http.MethodPost {
			// ~ TeamMemberRoleSpectator can't add members
			member, err := team_member_service.GetByMemberId(teamUUID, userUUID)
			if err != nil {
				utils.AbortRequest(w, "An error occured", http.StatusInternalServerError)
				return
			}
			if member.Role == models.TeamMemberRoleSpectator {
				utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
				return
			}
		}

		// ~ OK. Serve.
		next.ServeHTTP(w, r)
	})
}

// ~ /teams/{id}/members/{id} ~

func TeamMemberRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ~ Let's check that the user is authenticated
		if !auth_utils.CheckAuthenticationHeader(w, r) {
			return
		}

		userUUID, err := utils.ExtractUserIDFromJWT(r)
		if err != nil {
			utils.AbortRequest(w, "Authorization Token is invalid.", http.StatusUnauthorized)
			return
		}

		teamUUID, err := getTeamUUIDFromRequest(w, r)
		if err != nil {
			return
		}

		vars := mux.Vars(r)
		memberID := vars["memberID"]

		// ~ Check if the user is a member of the team
		isIn, err := team_service.IsUserInTeam(userUUID, teamUUID)
		if err != nil {
			utils.AbortRequest(w, "An error occured", http.StatusInternalServerError)
			return
		}
		if !isIn {
			utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// ~ Check team member permissions
		if r.Method == http.MethodDelete {
			// ~ TeamMemberRoleSpectator can't remove members
			member, err := team_member_service.GetByMemberId(teamUUID, userUUID)
			if err != nil {
				utils.AbortRequest(w, fmt.Sprintf("An error occured: %v", err), http.StatusInternalServerError)
				return
			}
			if member.Role == models.TeamMemberRoleSpectator {
				utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
				return
			}
		} else if r.Method == http.MethodPatch {
			// ~ TeamMemberRoleSpectator can't update members
			member, err := team_member_service.GetByMemberId(teamUUID, userUUID)
			if err != nil {
				utils.AbortRequest(w, "An error occured", http.StatusInternalServerError)
				return
			}
			if member.Role == models.TeamMemberRoleSpectator {
				utils.AbortRequest(w, "Unauthorized", http.StatusForbidden)
				return
			}

			// ~ Does member exist?
			_, err = team_member_service.Get(teamUUID, memberID)
			if err != nil {
				utils.AbortRequest(w, "Member not found", http.StatusNotFound)
				return
			}
		}

		// ~ OK. Serve.
		next.ServeHTTP(w, r)
	})
}
