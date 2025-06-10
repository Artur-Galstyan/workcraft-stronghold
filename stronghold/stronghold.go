package stronghold

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Artur-Galstyan/workcraft-stronghold/handlers"
	"github.com/Artur-Galstyan/workcraft-stronghold/static"
	"github.com/Artur-Galstyan/workcraft-stronghold/views"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"
)

type Stronghold struct {
	hashedAPIKey string
	client       *redis.Client
}

func NewStronghold(apiKey string, client *redis.Client) *Stronghold {

	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	hashedAPIKey := hex.EncodeToString(hasher.Sum(nil))

	return &Stronghold{
		hashedAPIKey: hashedAPIKey,
		client:       client,
	}

}

func (s *Stronghold) SetupBackgroundTasks() {
	var mutex sync.Mutex

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mutex.Lock()

			mutex.Unlock()

		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			mutex.Lock()
			mutex.Unlock()
		}
	}()
}

func (s *Stronghold) Run() {
	go s.SetupBackgroundTasks()
	s.StartHTTPServer()
}

func (s *Stronghold) StartHTTPServer() {
	fs := http.FileServer(static.GetFileSystem())
	http.Handle("/static/css/", http.StripPrefix("/static/css/", fs))

	component := views.Index()
	http.Handle("/", templ.Handler(component))
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			component := views.Login()
			templ.Handler(component).ServeHTTP(w, r)
		case http.MethodPost:
			handlers.LoginHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("GET /task/{id}", handlers.AuthMiddleware(handlers.CreateTaskViewHandler(), s.hashedAPIKey))
	http.HandleFunc("GET /tasks/", handlers.AuthMiddleware(handlers.TaskView, s.hashedAPIKey))
	http.HandleFunc("GET /peon/{id}/", handlers.AuthMiddleware(handlers.CreatePeonViewHandler(), s.hashedAPIKey))
	http.HandleFunc("GET /peons/", handlers.AuthMiddleware(handlers.CreatePeonsViewHandler(), s.hashedAPIKey))

	http.HandleFunc("GET /api/peons", handlers.AuthMiddleware(handlers.CreateGetPeonsHandler(), s.hashedAPIKey))
	http.HandleFunc("GET /api/peon/{id}", handlers.AuthMiddleware(handlers.CreateGetPeonHandler(), s.hashedAPIKey))
	http.HandleFunc("GET /api/peon/{id}/tasks", handlers.AuthMiddleware(handlers.CreateGetPeonTaskHandler(), s.hashedAPIKey))
	http.HandleFunc("POST /api/peon/{id}/update", handlers.AuthMiddleware(handlers.CreateUpdatePeonHandler(), s.hashedAPIKey))

	http.HandleFunc("POST /api/task", handlers.AuthMiddleware(handlers.CreatePostTaskHandler(s.client), s.hashedAPIKey))
	http.HandleFunc("GET /api/tasks", handlers.AuthMiddleware(handlers.CreateGetTasksHandler(), s.hashedAPIKey))
	http.HandleFunc("GET /api/task/{id}", handlers.AuthMiddleware(handlers.CreateGetTaskHandler(), s.hashedAPIKey))
	http.HandleFunc("POST /api/task/{id}/cancel", handlers.AuthMiddleware(handlers.CreateCancelTaskHandler(s.client), s.hashedAPIKey))
	http.HandleFunc("POST /api/task/{id}/update", handlers.AuthMiddleware(handlers.CreateTaskUpdateHandler(), s.hashedAPIKey))

	http.HandleFunc("GET /api/heartbeat", heartbeatHandler())

	slog.Info("Building Stronghold on port 6112")
	if err := http.ListenAndServe(":6112", nil); err != nil {
		slog.Error("Server failed", "error", err)
	}

}

func heartbeatHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET /heartbeat")
		w.Write([]byte("Success!"))
	}
}
