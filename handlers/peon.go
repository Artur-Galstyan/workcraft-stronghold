package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Artur-Galstyan/workcraft-stronghold/models"
	"github.com/Artur-Galstyan/workcraft-stronghold/utils"
	"github.com/Artur-Galstyan/workcraft-stronghold/views"
	"github.com/a-h/templ"
)

func CreateGetPeonsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams, err := utils.ParsePeonQuery(r.URL.Query().Get("query"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid query: %v", err), http.StatusBadRequest)
			return
		}
		fmt.Println(queryParams)
		w.Header().Set("Content-Type", "application/json")
	}
}

func CreateGetPeonHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peonID := r.PathValue("id")
		if peonID == "" {
			slog.Error("Peon ID is required")
			http.Error(w, "Peon ID is required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

func CreateGetPeonTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peonID := r.PathValue("id")
		if peonID == "" {
			slog.Error("Peon ID is required")
			http.Error(w, "Peon ID is required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

func CreateUpdatePeonHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peonID := r.PathValue("id")
		if peonID == "" {
			slog.Error("Peon ID is required")
			http.Error(w, "Peon ID is required", http.StatusBadRequest)
			return
		}
		var rawJSON map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&rawJSON); err != nil {
			slog.Error("Failed to decode request body", "err", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var update models.PeonUpdate
		jsonBytes, _ := json.Marshal(rawJSON)
		if err := json.Unmarshal(jsonBytes, &update); err != nil {
			slog.Error("Failed to parse update data", "err", err)
			http.Error(w, "Invalid update data", http.StatusBadRequest)
			return
		}

		_, update.StatusSet = rawJSON["status"]
		_, update.HeartbeatSet = rawJSON["last_heartbeat"]
		_, update.CurrentTaskSet = rawJSON["current_task"]
		_, update.QueuesSet = rawJSON["queues"]

		msg := fmt.Sprintf(`{"type": "peon_update", "message": {"peon": %s}}`,
			string(""))
		fmt.Println(msg)
		w.Header().Set("Content-Type", "application/json")
	}
}

func CreatePeonViewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peonID := r.PathValue("id")
		if peonID == "" {
			slog.Error("Peon ID is required")
			http.Error(w, "Peon ID is required", http.StatusBadRequest)
			return
		}
		component := views.Peon(peonID)
		templ.Handler(component).ServeHTTP(w, r)
	}
}

func CreatePeonsViewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		component := views.Peons()
		templ.Handler(component).ServeHTTP(w, r)
	}
}
