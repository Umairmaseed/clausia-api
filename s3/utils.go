package s3

import "fmt"

func GetPathToFile(filename, bucketName string) string {
	return fmt.Sprintf("s3://%s/%s", bucketName, filename)
}
