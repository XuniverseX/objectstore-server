package minio

import (
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	cfg "objectstore-server/config"
)

var minioClient *minio.Client

// Client 创建minio client对象
func Client() *minio.Client {
	// Initialize minio client object.
	c, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		fmt.Println(err)
	}

	minioClient = c
	return minioClient
}
