package server

import (
	"fmt"
	"gox/database"
	"gox/database/models"
	"gox/routes/auth"
	"gox/routes/teams"
	"gox/routes/users"
	"gox/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func createRoute(router *mux.Router, methods []string, route string, handler http.HandlerFunc, middlewares []func(http.Handler) http.Handler) {
	utils.ConsoleLog("🚦 Creating route %s %s", methods, route)

	middlewares = append(middlewares, RequestLoggerMiddleware)

	r := router.HandleFunc(route, handler).Methods(methods...)
	for _, middleware := range middlewares {
		r = r.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)
		}))
	}
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
			teams.HandleViewTeams(w, r)
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
	}, []func(http.Handler) http.Handler{teams.TeamMemberRouteMiddleware})

	createRoute(router, []string{http.MethodGet, http.MethodPatch, http.MethodDelete}, "/teams/{id}/members/{memberID}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teams.HandleGetTeamMember(w, r)
		} else if r.Method == http.MethodPatch {
			teams.HandleUpdateTeamMemberRole(w, r)

		} else if r.Method == http.MethodDelete {
			teams.HandleRemoveTeamMember(w, r)
		}
	}, []func(http.Handler) http.Handler{teams.TeamMemberRouteMiddleware})

	// ~ all others routes, 404
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 - Route Not Found", http.StatusNotFound)
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
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			authUserID = id
		}

		// Capture de la réponse HTTP (pour connaître le statut)
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Création de l'entrée de log
		logEntry := models.RequestLog{
			Domain:    r.Host,
			Endpoint:  r.URL.Path,
			Method:    r.Method,
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
