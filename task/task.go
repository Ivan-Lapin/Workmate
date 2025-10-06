package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Status — статус задачи
type Status string

const (
	StatusCompleted  Status = "завершена"
	StatusProcessing Status = "в работе"
)

// Task — описывает задачу по загрузке файлов
type Task struct {
	ID        string    `json:"id"`
	URLs      []string  `json:"urls"`
	Status    Status    `json:"status"`
	Errors    []error   `json:"errors"`
	CreatedAt time.Time `json:"created"`
	UpdatedAt time.Time `json:"updated"`
	mutex     sync.Mutex
}

// Save сохраняет задачу в файл в JSON
func (t *Task) Save(storageDir string) error {

	data, err := json.MarshalIndent(t, "", "")
	if err != nil {
		return fmt.Errorf("ошибка сериализации задачи: %v", err)
	}

	path := filepath.Join(storageDir, t.ID+"json")
	return os.WriteFile(path, data, 0644)
}

func LoadTask(storageDir, id string) (*Task, error) {
	path := filepath.Join(storageDir, id+"json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла задачи: %v", err)
	}

	var task Task
	if err = json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("ошибка десериализации задачи: %v", err)
	}

	return &task, nil
}
