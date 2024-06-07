package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifySignedPDF(t *testing.T) {
	filename := "../test/gosign-test-signed.pdf"

	file, err := io.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	r := bytes.NewReader(file)

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormFile("files", "test.pdf"); err != nil {
		t.FailNow()
	}

	if _, err = io.Copy(fw, r); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/verifydocs", &b)
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

	type verifyResponse struct {
		Name  string `json:"name"`
		Valid bool   `json:"valid"`
		Error string `json:"error"`
	}

	var resObject struct {
		Files []verifyResponse `json:"files"`
	}

	err = json.Unmarshal(resBytes, &resObject)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	responseList := resObject.Files
	for _, v := range responseList {
		if len(v.Name) == 0 {
			t.Error("response filename is empty")
		}

		if v.Valid != true {
			t.Error("signature not valid")
		}

		if v.Error != "" {
			t.Error(v.Error)
		}
	}
}

func TestVerifyWrongFilesForm(t *testing.T) {
	filename := "../test/gosign-test-signed.pdf"

	file, err := io.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	r := bytes.NewReader(file)

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormFile("filessss", "test.pdf"); err != nil {
		t.FailNow()
	}

	if _, err = io.Copy(fw, r); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/verifydocs", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}
}
