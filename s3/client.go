package s3

import (
	"bytes"
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

type S3Client struct {
	*s3.Client
}

func NewS3Client() (*S3Client, error) {
	conf, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedCredentialsFiles([]string{os.Getenv("S3_CREDENTIALS")}),
		config.WithSharedConfigProfile(os.Getenv("S3_PROFILE")),
		config.WithRegion(os.Getenv("S3_REGION")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(conf)

	return &S3Client{client}, nil
}

func (c *S3Client) UploadDocument(file []byte, filename, bucketName string) error {
	_, err := c.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(file),
	})

	return err
}
