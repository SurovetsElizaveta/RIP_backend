package minio

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func (c *Client) UploadImage(file *multipart.FileHeader, serviceID int) (string, error) {
	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer src.Close()

	// Генерируем уникальное имя файла на латинице
	ext := filepath.Ext(file.Filename)
	fileName := generateFileName(serviceID, ext)

	// Загружаем в Minio
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := c.client.PutObject(ctx, c.config.Bucket, fileName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки файла в Minio: %w", err)
	}

	// Формируем URL для доступа к файлу
	imageURL := c.generateFileURL(fileName)

	log.Printf("Файл загружен: %s, размер: %d байт", fileName, info.Size)
	return imageURL, nil
}

func (c *Client) DeleteImage(imageURL string) error {
	// Извлекаем имя файла из URL
	fileName := extractFileNameFromURL(imageURL)
	if fileName == "" {
		return fmt.Errorf("неверный URL изображения")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.client.RemoveObject(ctx, c.config.Bucket, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка удаления файла из Minio: %w", err)
	}

	log.Printf("Файл удален: %s", fileName)
	return nil
}

func generateFileName(serviceID int, extension string) string {
	// Убираем точку из расширения если есть
	ext := strings.TrimPrefix(extension, ".")
	if ext != "" {
		ext = "." + ext
	}

	// Генерируем уникальный UUID
	uniqueID := uuid.New().String()

	// Формируем имя: service_{id}_{uuid}{.ext}
	fileName := fmt.Sprintf("service_%d_%s%s", serviceID, uniqueID, ext)

	return strings.ToLower(fileName)
}

func (c *Client) generateFileURL(fileName string) string {
	protocol := "http"
	if c.config.UseSSL {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", protocol, c.config.Endpoint, c.config.Bucket, fileName)
}

func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (c *Client) GetFileInfo(fileName string) (*minio.ObjectInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stat, err := c.client.StatObject(ctx, c.config.Bucket, fileName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о файле: %w", err)
	}

	return &stat, nil
}
