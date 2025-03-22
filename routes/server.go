package server

import (
	"bytes"
	"fmt"
	"gox/database"
	"gox/database/models"
	"gox/routes/administration"
	admin_auth "gox/routes/administration/auth"
	admin_logs "gox/routes/administration/logs"
	admin_subscriptions "gox/routes/administration/subscriptions"
	"gox/routes/auth"
	"gox/routes/teams"
	"gox/routes/users"
	"gox/utils"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func createRoute(router *mux.Router, methods []string, route string, handler http.HandlerFunc, middlewares []func(http.Handler) http.Handler) {
	utils.ConsoleLog("🚦 Creating route %s %s", methods, route)

	// Add RequestLoggerMiddleware to all routes, must be the first middleware
	finalHandler := http.Handler(http.HandlerFunc(handler))
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	router.Handle(route, finalHandler).Methods(methods...)
}

func Start() {
	utils.ConsoleLog("🚀 Starting server...")

	host := utils.GetEnv("SERVER_HOST", "localhost")
	port := utils.GetEnv("SERVER_PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	router := mux.NewRouter()

	createRoute(router, []string{http.MethodGet}, "/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}, nil)

	// ~ AUTH ~

	createRoute(router, []string{http.MethodPost}, "/auth/login", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleLogin(w, r)
	}, nil)

	createRoute(router, []string{http.MethodPost}, "/auth/register", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleRegister(w, r)
	}, nil)

	// ~ USERS ~

	createRoute(router, []string{http.MethodGet, http.MethodPost}, "/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUsers(w, r)
		} else if r.Method == http.MethodPost {
			users.HandleCreateUser(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UsersRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPatch, http.MethodDelete}, "/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUser(w, r)
		} else if r.Method == http.MethodPatch {
			users.HandleUpdateUser(w, r)
		} else if r.Method == http.MethodDelete {
			users.HandleDeleteUser(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UserRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPost}, "/users/{id}/teams", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUserTeams(w, r)

		} else if r.Method == http.MethodPost {
			teams.HandleCreateTeam(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UserRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete}, "/users/{id}/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUserProfile(w, r)
		} else if r.Method == http.MethodPost {
			users.HandleCreateUserProfile(w, r)
		} else if r.Method == http.MethodPatch {
			users.HandleUpdateUserProfile(w, r)
		} else if r.Method == http.MethodDelete {
			users.HandleDeleteUserProfile(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UserRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPost}, "/users/{id}/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUserSubscriptions(w, r)
		} else if r.Method == http.MethodPost {
			users.HandleCreateUserSubscription(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UserRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPatch, http.MethodDelete}, "/users/{id}/subscriptions/{subscription_id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUserSubscription(w, r)
		} else if r.Method == http.MethodPatch {
			users.HandleUpdateUserSubscription(w, r)
		} else if r.Method == http.MethodDelete {
			users.HandleDeleteUserSubscription(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UserRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPatch}, "/users/{id}/subscriptions/{subscription_id}/perks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users.HandleGetUserSubscriptionPerks(w, r)
		} else if r.Method == http.MethodPatch {
			users.HandleUpdateUserSubscriptionPerks(w, r)
		}
	}, []func(http.Handler) http.Handler{users.UserRouteMiddleware})

	// ~ TEAMS ~

	createRoute(router, []string{http.MethodGet, http.MethodPost}, "/teams", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teams.HandleViewTeams(w, r)
		} else if r.Method == http.MethodPost {
			teams.HandleCreateTeam(w, r)
		}
	}, []func(http.Handler) http.Handler{teams.TeamsRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPatch, http.MethodDelete}, "/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teams.HandleGetTeam(w, r)
		} else if r.Method == http.MethodPatch {
			teams.HandleUpdateTeam(w, r)
		} else if r.Method == http.MethodDelete {
			teams.HandleDeleteTeam(w, r)
		}
	}, []func(http.Handler) http.Handler{teams.TeamRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPost}, "/teams/{id}/members", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teams.HandleGetTeamMembers(w, r)
		} else if r.Method == http.MethodPost {
			teams.HandleAddTeamMember(w, r)
		}
	}, []func(http.Handler) http.Handler{teams.TeamRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPatch, http.MethodDelete}, "/teams/{id}/members/{member_id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teams.HandleGetTeamMember(w, r)
		} else if r.Method == http.MethodPatch {
			teams.HandleUpdateTeamMemberRole(w, r)
		} else if r.Method == http.MethodDelete {
			teams.HandleRemoveTeamMember(w, r)
		}
	}, []func(http.Handler) http.Handler{teams.TeamMemberRouteMiddleware})

	// ~ ADMINISTRATION ~

	createRoute(router, []string{http.MethodPost}, "/administrate/login", func(w http.ResponseWriter, r *http.Request) {
		admin_auth.HandleLogin(w, r)
	}, nil)

	createRoute(router, []string{http.MethodGet, http.MethodPost}, "/administrate/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin_subscriptions.HandleGetSubscriptions(w, r)
		} else if r.Method == http.MethodPost {
			admin_subscriptions.HandleCreateSubscription(w, r)
		}
	}, []func(http.Handler) http.Handler{administration.AdministrationRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete}, "/administrate/subscriptions/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin_subscriptions.HandleGetSubscription(w, r)
		} else if r.Method == http.MethodPatch {
			admin_subscriptions.HandleUpdateSubscription(w, r)
		} else if r.Method == http.MethodDelete {
			admin_subscriptions.HandleDeleteSubscription(w, r)
		}
	}, []func(http.Handler) http.Handler{administration.AdministrationRouteMiddleware})

	createRoute(router, []string{http.MethodGet}, "/administrate/logs", func(w http.ResponseWriter, r *http.Request) {
		admin_logs.HandleGetLogs(w, r)
	}, []func(http.Handler) http.Handler{administration.AdministrationRouteMiddleware})

	// ~ all others routes, 404
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.AbortRequest(w, "404 - Route Not Found", http.StatusNotFound)
	})

	utils.ConsoleLog("🌍 Server started on http://%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		utils.ConsoleLog("❌ Server failed to start: %v", err).Fatal()
	}

}

// responseRecorder intercepte le statut HTTP
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader met à jour le statusCode et appelle la méthode WriteHeader de http.ResponseWriter
func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Récupérer l'ID utilisateur depuis le JWT Token
		tokenString := r.Header.Get("Authorization")
		var authUserID uuid.UUID = uuid.Nil

		if tokenString != "" {
			id, err := utils.ExtractUserIDFromJWT(r)
			if err != nil {
				utils.ConsoleLog("❌ Erreur lors de la récupération de l'ID utilisateur: %v", err)
				utils.AbortRequest(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			authUserID = id
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			utils.ConsoleLog("❌ Erreur lors de la lecture du corps de la requête: %v", err)
			utils.AbortRequest(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Encode the body for logging
		encodedBody := utils.EncodeBase64(body)

		// Restore the body for downstream handlers
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Capture de la réponse HTTP (pour connaître le statut)
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Création de l'entrée de log
		logEntry := models.RequestLog{
			Domain:    r.Host,
			Endpoint:  r.URL.Path,
			Method:    r.Method,
			Content:   string(encodedBody),
			Status:    rec.statusCode,
			Timestamp: start,
		}

		// Ajout du UserID si ce n'est pas un utilisateur anonyme
		if authUserID != uuid.Nil {
			logEntry.UserID = &authUserID
			utils.ConsoleLog("📜 Log enregistré -> [%s] %s %s -> %d", authUserID, r.Method, r.URL.Path, rec.statusCode)
		} else {
			utils.ConsoleLog("📜 Log enregistré -> [anonymous] %s %s -> %d", r.Method, r.URL.Path, rec.statusCode)
		}

		// Enregistrement du log en base de données
		if err := database.DB.Create(&logEntry).Error; err != nil {
			utils.ConsoleLog("⚠️ Erreur lors de l'enregistrement du log en DB: %v", err)
		}
	})
}
