package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"test_task/url_downloader/task"
)

// Server — HTTP-сервер, принимающий задачи
type Server struct {
	taskManager *task.TaskManager
	server      *http.Server
}

// NewServer создаёт новый HTTP-сервер
func NewServer(taskManager *task.TaskManager) *Server {
	srv := &Server{
		taskManager: taskManager,
		server: &http.Server{
			Addr: ":8089",
		},
	}

	// Регистрация обработчиков
	http.HandleFunc("/create", srv.handleCreate)
	http.HandleFunc("/status", srv.handleStatus)

	return srv
}

// handleCreate — обработчик POST /create
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "только POST запросы разрешены", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URLs []string `json:"urls"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "список URL не может быть пустым", http.StatusBadRequest)
		return
	}

	urlRegex := regexp.MustCompile(`^https?://`)
	for _, url := range req.URLs {
		if !urlRegex.MatchString(url) {
			http.Error(w, "некорректный URL: "+url, http.StatusBadRequest)
			return
		}
	}

	// Создаём задачу
	taskID := s.taskManager.CreateTask(req.URLs)

	// Ответ: ID новой задачи
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id": taskID,
	})

}

// handleStatus — обработчик GET /status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "только GET запросы разрешены", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "требуется параметр id", http.StatusBadRequest)
		return
	}

	// Получаем задачу
	task, exists := s.taskManager.GetTask(id)
	if !exists {
		http.Error(w, "задача не найдена", http.StatusNotFound)
		return
	}

	// Возвращаем статус в JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
