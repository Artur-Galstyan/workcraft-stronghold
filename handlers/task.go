package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Artur-Galstyan/workcraft-stronghold/events"
	"github.com/Artur-Galstyan/workcraft-stronghold/models"
	"github.com/Artur-Galstyan/workcraft-stronghold/sqls"
	"github.com/Artur-Galstyan/workcraft-stronghold/utils"
	"github.com/Artur-Galstyan/workcraft-stronghold/views"
	"github.com/a-h/templ"
	"gorm.io/gorm"
)

func CreateTaskViewHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}

		task, err := sqls.GetTask(db, taskID)
		if err != nil {
			slog.Error("Error querying task", "err", err)
			http.Error(w, "Unable to find task", http.StatusNotFound)
			return
		}

		component := views.Task(task)
		templ.Handler(component).ServeHTTP(w, r)
	}
}

func TaskView(w http.ResponseWriter, r *http.Request) {
	component := views.Tasks()
	templ.Handler(component).ServeHTTP(w, r)
}

func CreateTaskUpdateHandler(db *gorm.DB, eventSender *events.EventSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}
		slog.Info("Received update for task", "id", taskID)

		var update models.TaskUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			slog.Error("Failed to decode request body", "err", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updatedTask, err := sqls.UpdateTask(db, taskID, update)

		taskJSON, err := json.Marshal(updatedTask)
		if err != nil {
			slog.Error("Failed to serialize updated task", "err", err)
		} else {
			msg := fmt.Sprintf(`{"type": "task_update", "message": {"task": %s}}`,
				string(taskJSON))
			eventSender.BroadcastToChieftains(msg)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
			slog.Error("Failed to encode response", "err", err)
		}
	}
}

func CreatePostTaskHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received new task!")

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			slog.Error("Failed to decode request body", "err", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		task, err := sqls.CreateTask(db, task)
		if err != nil {
			slog.Error("Failed to create task", "err", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(task); err != nil {
			slog.Error("Failed to encode task", "err", err)
		}
	}
}

func CreateGetTasksHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET /api/tasks")
		queryString := r.URL.Query().Get("query")

		queryParams, err := utils.ParseTaskQuery(queryString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid query: %v", err), http.StatusBadRequest)
			return
		}

		response, err := sqls.GetTasks(db, *queryParams)
		if err != nil {
			slog.Error("Failed to fetch tasks", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("Error encoding response", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func CreateGetTaskHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET /api/task/{id}")

		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}

		slog.Info("Fetching task", "id", taskID)

		task, err := sqls.GetTask(db, taskID)
		if err != nil {
			slog.Error("Failed to fetch task", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(task); err != nil {
			slog.Error("Failed to encode task", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func CreateCancelTaskHandler(db *gorm.DB, eventSender *events.EventSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("POST /api/task/{id}/cancel")

		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}
		slog.Info("Received cancel for task", "id", taskID)

		task, err := sqls.GetTask(db, taskID)
		if err != nil {
			slog.Error("Failed to fetch task", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		msg := fmt.Sprintf(`{"type": "cancel_task", "data": "%s"}`, taskID)
		if err := eventSender.SendEvent(*task.PeonID, msg); err != nil {
			slog.Error("Failed to send cancel message", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		res := map[string]interface{}{
			"send": "true",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("Failed to encode response", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
