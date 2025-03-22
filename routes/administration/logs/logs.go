package admin_logs

import (
	admin_logs_service "gox/services/administration/logs"
	"gox/utils"
	"net/http"
)

func HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := admin_logs_service.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, logs)
}
