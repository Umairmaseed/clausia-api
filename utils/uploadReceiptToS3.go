package utils

import (
	"os"

	"github.com/goledgerdev/goprocess-api/s3"
	"github.com/google/logger"
)

func UploadReceiptToS3(file []byte, fileName string) (string, error) {
	s3Client, err := s3.NewS3Client()
	if err != nil {
		logger.Error(err)

		return "", err
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	filename := "receipts/" + fileName
	err = s3Client.UploadDocument(file, filename, bucketName)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	return s3.GetPathToFile(filename, bucketName), nil
}
