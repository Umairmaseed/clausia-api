package utils

import (
	"io"
	"mime/multipart"
)

// Get bytes from a file
func GetFileBytes(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
