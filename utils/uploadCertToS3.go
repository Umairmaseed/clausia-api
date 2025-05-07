package utils

import (
	"os"

	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/s3"
)

func UploadCertToS3(cert []byte, certName string) (string, error) {
	s3Client, err := s3.NewS3Client()
	if err != nil {
		logger.Error(err)
		return "", err
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")

	filename := "certificates/" + certName

	err = s3Client.UploadDocument(cert, filename, bucketName)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	return s3.GetPathToFile(filename, bucketName), nil
}
