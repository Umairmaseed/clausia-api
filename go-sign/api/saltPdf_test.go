package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestSaltPdf(t *testing.T) {

	filename := "../test/gosign-test-document.pdf"

	file, err := io.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	r := bytes.NewReader(file)

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer
	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, r); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/saltPdf", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("status: ", w.Code)
		t.FailNow()
	}

	resBytes, err := io.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var resObject struct {
		File       []byte `json:"file"`
		SaltedHash string `json:"saltedHash"`
	}

	err = json.Unmarshal(resBytes, &resObject)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	saltedFile := resObject.File
	uuidBytes := saltedFile[len(saltedFile)-36:]

	// Check if a uuid is in the last 36 bytes of the PDF
	_, err = uuid.ParseBytes(uuidBytes)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestSaltPdfWrongRequest(t *testing.T) {

	filename := "../test/gosign-test-document.pdf"

	file, err := io.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	r := bytes.NewReader(file)

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer
	if fw, err = multipartWriter.CreateFormFile("filee", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, r); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/saltPdf", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}
}

// TODO

// func TestSaltPdfWrongFile(t *testing.T) {
// 	filename := "../test/gosign-test-document-protected.pdf"

// 	file, err := io.ReadFile(filename)
// 	if err != nil {
// 		t.Errorf("failed to read file")
// 		t.FailNow()
// 	}

// 	r := bytes.NewReader(file)

// 	var b bytes.Buffer
// 	multipartWriter := multipart.NewWriter(&b)
// 	var fw io.Writer
// 	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
// 		t.FailNow()
// 	}
// 	if _, err = io.Copy(fw, r); err != nil {
// 		t.FailNow()
// 	}

// 	multipartWriter.Close()

// 	w := httptest.NewRecorder()
// 	req, err := http.NewRequest("POST", "/api/saltPdf", &b)
// 	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
// 	router.ServeHTTP(w, req)

// 	if w.Code != 400 {
// 		t.Error(err)
// 		t.FailNow()
// 	}
// }
