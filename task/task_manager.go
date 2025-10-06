package task

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"test_task/url_downloader/storage"
	"time"
)

// TaskManager — центральный объект управления задачами сккряяяяя
type TaskManager struct {
	taskDir   string
	tasks     map[string]*Task
	mutex     sync.Mutex
	wg        sync.WaitGroup
	ctx       context.Context
	workerNum int
	fm        *storage.FileManager
}

// NewTaskManager создаёт и инициализирует менеджер задач
func NewTaskManager(ctx context.Context, taskDir, downloadDir string, workerNum int) *TaskManager {

	fm := storage.NewFileManager(downloadDir)

	if err := os.MkdirAll(taskDir, 0755); err != nil {
		panic(fmt.Sprintf("не удалось создать директорию задач: %v", err))
	}

	tm := &TaskManager{
		taskDir:   taskDir,
		tasks:     make(map[string]*Task),
		ctx:       ctx,
		workerNum: workerNum,
		fm:        fm,
	}

	if err := tm.loadTasks(); err != nil {
		panic(fmt.Sprintf("не удалось загрузить задачи: %v", err))
	}

	// Запуск воркеров
	tm.startWorkers()

	return tm
}

func (tm *TaskManager) CreateTask(urls []string) string {
	log.Printf("🔧 CreateTask: получено %d URL\n", len(urls)) // ⬅ Добавь

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	id := generateID()

	task := &Task{
		ID:        id,
		URLs:      urls,
		Status:    StatusProcessing,
		Errors:    nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tm.tasks[id] = task
	log.Printf("Задача %s добавлена в мапу\n", id) // ⬅

	// Сохраняем
	if err := task.Save(tm.taskDir); err != nil {
		log.Printf("Ошибка сохранения задачи: %v", err)
	} else {
		log.Printf("Задача %s сохранена в %s\n", id, tm.taskDir)
	}

	// Запускаем обработку
	go tm.startProcessing(task)

	return id
}

// GetTask возвращает задачу по ID
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	task, exist := tm.tasks[id]
	return task, exist
}

func (tm *TaskManager) loadTasks() error {

	files, err := os.ReadDir(tm.taskDir)
	if err != nil {
		return fmt.Errorf("Неудачное считывание директории с ошибкой: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			id := file.Name()[:len(file.Name())-5]
			task, err := LoadTask(tm.taskDir, id)
			if err != nil {
				log.Printf("Ошибка загрузки задачи %s: %v", id, err)
				continue
			}
			tm.tasks[id] = task

			// Если задача была в работе — перезапускаем обработку
			if task.Status == StatusProcessing {
				tm.startProcessing(task)
			}
		}
	}

	return nil
}

// GenerateID — генератор ID на основе временной метки
func generateID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// startWorkers запускает пул воркеров
func (tm *TaskManager) startWorkers() {
	for i := 0; i < tm.workerNum; i++ {
		tm.wg.Add(1)
		go tm.worker()
	}
}

// worker — бесконечный цикл, который ждёт задачи (в дальнейшем будет через канал)
// Пока оставим пустым — реализуем в следующем шаге с `storage`
func (tm *TaskManager) worker() {
	defer tm.wg.Done()

	for {
		select {
		case <-tm.ctx.Done():
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// startProcessing — запускает загрузку всех URL из задачи
func (tm *TaskManager) startProcessing(task *Task) {
	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()

		for _, url := range task.URLs {
			select {
			case <-tm.ctx.Done():
				return
			default:
			}

			filename, err := tm.fm.DownloadFile(tm.ctx, url, task.ID)
			if err != nil {
				task.mutex.Lock()
				task.Errors = append(task.Errors, fmt.Errorf("ошибка загрузки %s: %v", url, err))
				task.mutex.Unlock()
				continue
			}

			log.Printf("Загружен файл: %s для задачи %s", filename, task.ID)
		}

		task.mutex.Lock()
		task.Status = StatusCompleted
		task.UpdatedAt = time.Now()
		task.Save(tm.taskDir)
		task.mutex.Unlock()
	}()
}
