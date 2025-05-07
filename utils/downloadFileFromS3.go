package utils

import (
	"context"
	"os"

	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/s3"
)

func DownloadFileFromS3(ctx context.Context, filename string) ([]byte, error) {

	s3Client, err := s3.NewS3Client()
	if err != nil {
		logger.Error("Failed to create new S3 client", err)
		return nil, err
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")

	docBytes, err := s3Client.DownloadDocument(ctx, filename, bucketName)
	if err != nil {
		logger.Error("Failed to download document", err)
		return nil, err
	}

	return docBytes, nil
}
