// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"test_task/url_downloader/server"
	"test_task/url_downloader/task"
	"time"
)

func main() {
	// Контекст для отмены всех долгоживущих операций (горутин)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// === Инициализация зависимостей ===
	// TaskManager создаёт FileManager внутри себя
	tm := task.NewTaskManager(ctx, "task_storage", "download_storage", 5)

	// Создаём сервер — он сразу регистрирует свои обработчики в DefaultServeMux
	server.NewServer(tm)

	// === Настройка HTTP-сервера ===
	httpServer := &http.Server{
		Addr:    ":8089",
		Handler: nil, // Использует DefaultServeMux (где уже есть /create и /status)
	}

	// Запуск HTTP-сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("HTTP сервер запущен на порту: %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP сервер остановлен с ошибкой: %v\n", err)
		}
	}()

	// === Перехват сигналов ОС: SIGINT, SIGTERM ===
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнала остановки
	log.Println("Сервис запущен. Для остановки нажмите Ctrl+C")
	<-signChan
	log.Println("Сервис остановлен. Выполняется graceful shutdown...")

	// === Грейсфул остановка ===
	// Отмена контекста — останавливаем все горутины
	cancel()

	// Грейсфул остановка HTTP-сервера
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Ошибка graceful shutdown HTTP-сервера: %v", err)
	}
	log.Println("HTTP сервер остановлен корректно.")

	// Ожидаем завершения всех горутин
	wg.Wait()
	log.Println("Сервис остановлен корректно.")
}
