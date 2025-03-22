package administration

import (
	auth_utils "gox/services/auth"
	"gox/utils"
	"net/http"
)

func AdministrationRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Vérifier que l’utilisateur est bien authentifié
		if !auth_utils.CheckAuthenticationHeader(w, r) {
			return
		}

		// Vérifier que l’utilisateur est bien un admin
		if !auth_utils.IsAuthenticatedUserAdmin(w, r) {
			utils.AbortRequest(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Passer à l’handler suivant
		next.ServeHTTP(w, r)
	})
}
