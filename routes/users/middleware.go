package users

import (
	auth_utils "gox/services/auth"
	team_service "gox/services/teams"
	"gox/utils"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// This middleware checks that the user is authenticated and has admin rights
func UsersRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// ~ Let's check that the user is authenticated
			if !auth_utils.CheckAuthenticationHeader(w, r) {
				return
			}

			// ~ Only admins can view all users
			if !auth_utils.IsAuthenticatedUserAdmin(w, r) {
				return
			}
		}

		// ~ OK. Serve.
		next.ServeHTTP(w, r)
	})
}

// This middleware checks that the user is authenticated and has rights to access the specified user
func UserRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ~ Let's check that the user is authenticated
		if !auth_utils.CheckAuthenticationHeader(w, r) {
			return
		}

		// ~ If the user is not an admin, let's check that the user has rights to access this specific user
		if auth_utils.IsAuthenticatedUserAdmin(w, r) {
			// ~ OK. Serve.
			next.ServeHTTP(w, r)
			return
		}

		// ~ Is it just accessing the user public profile?
		if strings.HasSuffix(r.URL.Path, "/profile") {
			// ~ OK. Serve.
			next.ServeHTTP(w, r)
			return
		}

		// ~ Let's check that the user has rights to access this specific user
		// ~ 1) is the user the same as the one in the URL?
		vars := mux.Vars(r)
		userID := vars["id"]
		if auth_utils.GetAuthenticatedUserID(w, r) == userID {
			// ~ OK. Serve.
			next.ServeHTTP(w, r)
			return
		} else if strings.HasSuffix(r.URL.Path, "/subscriptions") {
			http.Error(w, "User is not allowed to access this user's subscriptions", http.StatusForbidden)
			return
		}

		// ~ 2) is the requested user a member of the same team as the user in the URL?
		userUUID, err := utils.ExtractUserIDFromJWT(r)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			return
		}

		userTeams, err := team_service.GetTeamsByMemberID(userUUID)
		if err != nil {
			http.Error(w, "Error getting user teams", http.StatusInternalServerError)
			return
		}

		if len(userTeams) == 0 {
			http.Error(w, "User has no teams", http.StatusForbidden)
			return
		}

		userTeamsUUIDs := make([]uuid.UUID, len(userTeams))
		for i, team := range userTeams {
			userTeamsUUIDs[i] = team.ID
		}
		isIn := team_service.IsUserInTeams(userUUID, userTeamsUUIDs)
		if !isIn {
			http.Error(w, "User is not in a team with the user in the URL", http.StatusForbidden)
			return
		}

		// ~ OK. Serve.
		next.ServeHTTP(w, r)
	})
}
