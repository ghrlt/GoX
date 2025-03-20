package administration

import (
	"gox/utils"
	"net/http"
)

func AdministrationRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ConsoleLog("🔐 AdministrationRouteMiddleware")
		next.ServeHTTP(w, r)
	})
}
