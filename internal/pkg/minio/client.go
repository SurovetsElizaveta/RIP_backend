package minio

import (
	"context"
	"fmt"
	"log"
	"time"

	"rip/internal/app/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	config *config.MinioConfig
}

func NewMinioClient(cfg *config.MinioConfig) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Minio клиента: %w", err)
	}

	client := &Client{
		client: minioClient,
		config: cfg,
	}

	// Проверяем подключение и создаем бакет если нужно
	if err := client.initialize(); err != nil {
		return nil, err
	}

	return client, nil
}

// initialize проверяет подключение и создает бакет если его нет
func (c *Client) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем существует ли бакет
	exists, err := c.client.BucketExists(ctx, c.config.Bucket)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования бакета: %w", err)
	}

	if !exists {
		// Создаем бакет если не существует
		err = c.client.MakeBucket(ctx, c.config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("ошибка создания бакета: %w", err)
		}

		// Настраиваем политику доступа (публичный доступ для чтения)
		policy := `{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {"AWS": "*"},
                    "Action": ["s3:GetObject"],
                    "Resource": ["arn:aws:s3:::` + c.config.Bucket + `/*"]
                }
            ]
        }`

		err = c.client.SetBucketPolicy(ctx, c.config.Bucket, policy)
		if err != nil {
			log.Printf("Предупреждение: не удалось установить политику бакета: %v", err)
		}

		log.Printf("Бакет %s создан успешно", c.config.Bucket)
	}

	log.Printf("Minio клиент инициализирован. Бакет: %s", c.config.Bucket)
	return nil
}
