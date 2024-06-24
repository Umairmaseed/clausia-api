package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLedgerKey(t *testing.T) {
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
	req, _ := http.NewRequest("POST", "/api/getkey", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("status: ", w.Code)
		t.FailNow()
	}

	response, err := io.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var obj responseObject
	err = json.Unmarshal(response, &obj)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if fmt.Sprintf("%T", obj.LedgerKey) != "string" {
		t.Error(err)
		t.FailNow()
	}
}

func TestGetLedgerKeyWrongRequest(t *testing.T) {
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
	req, _ := http.NewRequest("POST", "/api/getkey", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}
}
