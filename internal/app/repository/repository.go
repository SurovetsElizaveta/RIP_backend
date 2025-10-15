package repository

import (
	"rip/internal/pkg/minio"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	minio *minio.Client
}

func New(dsn string, minioClient *minio.Client) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:    db,
		minio: minioClient,
	}, nil
}
