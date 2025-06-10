package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Artur-Galstyan/workcraft-stronghold/models"
	"github.com/Artur-Galstyan/workcraft-stronghold/utils"
	"github.com/Artur-Galstyan/workcraft-stronghold/views"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"
)

func CreateTaskViewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}
		task := models.Task{}
		component := views.Task(task)
		templ.Handler(component).ServeHTTP(w, r)
	}
}

func TaskView(w http.ResponseWriter, r *http.Request) {
	component := views.Tasks()
	templ.Handler(component).ServeHTTP(w, r)
}

func CreateTaskUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("POST /api/task/{id}/update")
		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}

		var rawJSON map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&rawJSON); err != nil {
			slog.Error("Failed to decode request body", "err", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		slog.Info("Received task update", "data", rawJSON)

		jsonBytes, err := json.Marshal(rawJSON)
		if err != nil {
			slog.Error("Failed to marshal JSON", "err", err)
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		var update models.TaskUpdate
		if err := json.Unmarshal(jsonBytes, &update); err != nil {
			slog.Error("Failed to parse update data", "err", err)
			http.Error(w, "Invalid update data", http.StatusBadRequest)
			return
		}

		// Set the flags based on field presence
		_, update.StatusSet = rawJSON["status"]
		_, update.TaskNameSet = rawJSON["task_name"]
		_, update.PeonIDSet = rawJSON["peon_id"]
		_, update.QueueSet = rawJSON["queue"]
		_, update.PayloadSet = rawJSON["payload"]
		_, update.ResultSet = rawJSON["result"]
		_, update.RetryOnFailureSet = rawJSON["retry_on_failure"]
		_, update.RetryCountSet = rawJSON["retry_count"]
		_, update.RetryLimitSet = rawJSON["retry_limit"]
		_, update.LogsSet = rawJSON["logs"]

		w.Header().Set("Content-Type", "application/json")
	}
}

func CreatePostTaskHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received new task!")

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			slog.Error("Failed to decode request body", "err", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		task.Status = models.TaskStatusPending

		payloadJSON, _ := json.Marshal(task.Payload)
		resultJSON, _ := json.Marshal(task.Result)

		taskDataForHash := map[string]interface{}{
			"id":               task.ID,
			"created_at":       task.CreatedAt.Format(time.RFC3339),
			"updated_at":       task.UpdatedAt.Format(time.RFC3339),
			"task_name":        task.TaskName,
			"status":           string(task.Status),
			"queue":            task.Queue,
			"retry_on_failure": task.RetryOnFailure,
			"payload":          string(payloadJSON),
			"result":           string(resultJSON),
		}

		ctx := context.Background()
		taskKey := "task:" + task.ID

		pipe := client.TxPipeline()
		pipe.HSet(ctx, taskKey, taskDataForHash)
		pipe.SAdd(ctx, "tasks:status:"+string(task.Status), task.ID)
		pipe.SAdd(ctx, "tasks:task_name:"+task.TaskName, task.ID)
		pipe.SAdd(ctx, "tasks:queue:"+task.Queue, task.ID)

		if _, err := pipe.Exec(ctx); err != nil {
			slog.Error("Failed to execute transaction to store task", "err", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		fullTaskJSON, _ := json.Marshal(task)
		client.LPush(ctx, task.Queue, fullTaskJSON)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

func CreateGetTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryString := r.URL.Query().Get("query")

		queryParams, err := utils.ParseTaskQuery(queryString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid query: %v", err), http.StatusBadRequest)
			return
		}
		fmt.Println(queryParams)

		w.Header().Set("Content-Type", "application/json")
	}
}

func CreateGetTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET /api/task/{id}")

		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}

		slog.Info("Fetching task", "id", taskID)
		w.Header().Set("Content-Type", "application/json")
	}
}

func CreateCancelTaskHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("POST /api/task/{id}/cancel")

		taskID := r.PathValue("id")
		if taskID == "" {
			slog.Error("Task ID is required")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}
		slog.Info("Received cancel for task", "id", taskID)

		ctx := context.Background()

		if err := client.Publish(ctx, "cancel", taskID).Err(); err != nil {
			slog.Error("Failed to publish cancellation message", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		res := map[string]interface{}{
			"cancellation_sent": "true",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("Failed to encode response", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
