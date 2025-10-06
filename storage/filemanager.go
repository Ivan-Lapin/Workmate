// storage/filemanager.go
package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FileManager — работает с файлами и URL
type FileManager struct {
	downloadDir string
	client      *http.Client
}

// NewFileManager создаёт новый FileManager
func NewFileManager(downloadDir string) *FileManager {
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		panic(fmt.Sprintf("не удалось создать директорию загрузки: %v", err))
	}

	return &FileManager{
		downloadDir: downloadDir,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DownloadFile загружает файл по URL и сохраняет в downloadDir
// Возвращает имя сохранённого файла или ошибку
func (fm *FileManager) DownloadFile(ctx context.Context, url, taskID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := fm.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	filename := fmt.Sprintf("%s_%s", taskID, filepath.Base(url))

	// Защита от пустого имени
	if filename == "" || filename == "." || filename == ".." {
		filename = fmt.Sprintf("%s_%d", taskID, time.Now().UnixNano())
	}
	filepath := filepath.Join(fm.downloadDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return filename, err
	}

	return filename, nil
}
