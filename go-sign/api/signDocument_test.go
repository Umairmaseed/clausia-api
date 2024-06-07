package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goledgerdev/go-sign/pdfsign"
)

func TestSignDocument(t *testing.T) {
	fileName := "../test/gosign-test-document.pdf"
	certificateName := "../test/47773901104.pfx"

	file, err := io.ReadFile(fileName)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	certificate, err := io.ReadFile(certificateName)
	if err != nil {
		t.Errorf("failed to read certificate")
		t.FailNow()
	}

	reqKey := "document:6e6ebec1-741f-535c-b70d-8145365606ad"

	fileNameReader := strings.NewReader("gosign-test-document.pdf")
	fileReader := bytes.NewReader(file)
	certificateReader := bytes.NewReader(certificate)
	passwordReader := strings.NewReader("123pmesp456")
	signatureReader := strings.NewReader("{ \"gosign-test-document.pdf\": { \"rect\": { \"x\": 309.609375, \"y\": 400.33333333333337, \"page\": 1 }, \"final\": false } }")
	ledgerKeyReader := strings.NewReader(reqKey)
	baseUrlReader := strings.NewReader("localhost/verify")

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormField("fileName"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileNameReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("certificate", "certificate.pfx"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, certificateReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("password"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, passwordReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("signature"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, signatureReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("ledgerKey"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, ledgerKeyReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("clientBaseUrl"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, baseUrlReader); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signdocs", &b)
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
		Name string `json:"name"`
		File []byte `json:"file"`
	}

	err = json.Unmarshal(resBytes, &resObject)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	responseFileName := resObject.Name
	if len(responseFileName) == 0 {
		t.Error("response filename is empty")
		t.FailNow()
	}

	responseFile := resObject.File
	key, err := pdfsign.RetrieveKey(responseFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expectedKey := strings.Split(reqKey, ":")[1]

	if key != expectedKey {
		t.Error("key is different from request key")
		t.Error("key: ", key)
		t.Error("request key: ", expectedKey)
		t.FailNow()
	}

	err = io.WriteFile("../test/sign-document-result.pdf", resObject.File, 0644)
}

func TestSignSignedDocument(t *testing.T) {
	fileName := "../test/gosign-test-signed.pdf"
	certificateName := "../test/47773901104.pfx"

	file, err := io.ReadFile(fileName)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	certificate, err := io.ReadFile(certificateName)
	if err != nil {
		t.Errorf("failed to read certificate")
		t.FailNow()
	}

	fileNameReader := strings.NewReader("gosign-test-signed.pdf")
	fileReader := bytes.NewReader(file)
	certificateReader := bytes.NewReader(certificate)
	passwordReader := strings.NewReader("123pmesp456")
	signatureReader := strings.NewReader("{ \"gosign-test-signed.pdf\": { \"rect\": { \"x\": 309.609375, \"y\": 900.33333333333337, \"page\": 1 }, \"final\": false } }")
	baseUrlReader := strings.NewReader("localhost/verify")

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormField("fileName"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileNameReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("certificate", "certificate.pfx"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, certificateReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("password"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, passwordReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("signature"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, signatureReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("clientBaseUrl"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, baseUrlReader); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signdocs", &b)
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
		Name string `json:"name"`
		File []byte `json:"file"`
	}

	err = json.Unmarshal(resBytes, &resObject)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	responseFileName := resObject.Name
	if len(responseFileName) == 0 {
		t.Error("response filename is empty")
		t.FailNow()
	}

	responseFile := resObject.File
	_, err = pdfsign.RetrieveKey(responseFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestSignWrongFileForm(t *testing.T) {
	fileName := "../test/gosign-test-document.pdf"
	certificateName := "../test/47773901104.pfx"

	file, err := io.ReadFile(fileName)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	certificate, err := io.ReadFile(certificateName)
	if err != nil {
		t.Errorf("failed to read certificate")
		t.FailNow()
	}

	fileNameReader := strings.NewReader("gosign-test-document.pdf")
	fileReader := bytes.NewReader(file)
	certificateReader := bytes.NewReader(certificate)
	passwordReader := strings.NewReader("123pmesp456")
	signatureReader := strings.NewReader("{ \"gosign-test-document.pdf\": { \"rect\": { \"x\": 309.609375, \"y\": 900.33333333333337, \"page\": 1 }, \"final\": false } }")
	baseUrlReader := strings.NewReader("localhost/verify")

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormField("fileName"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileNameReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("fileee", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("certificate", "certificate.pfx"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, certificateReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("password"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, passwordReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("signature"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, signatureReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("clientBaseUrl"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, baseUrlReader); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signdocs", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}

}

func TestSignNoKey(t *testing.T) {
	fileName := "../test/gosign-test-document.pdf"
	certificateName := "../test/47773901104.pfx"

	file, err := io.ReadFile(fileName)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	certificate, err := io.ReadFile(certificateName)
	if err != nil {
		t.Errorf("failed to read certificate")
		t.FailNow()
	}

	fileNameReader := strings.NewReader("gosign-test-document.pdf")
	fileReader := bytes.NewReader(file)
	certificateReader := bytes.NewReader(certificate)
	passwordReader := strings.NewReader("123pmesp456")
	signatureReader := strings.NewReader("{ \"gosign-test-document.pdf\": { \"rect\": { \"x\": 309.609375, \"y\": 900.33333333333337, \"page\": 1 }, \"final\": false } }")
	baseUrlReader := strings.NewReader("localhost/verify")

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormField("fileName"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileNameReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("certificate", "certificate.pfx"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, certificateReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("password"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, passwordReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("signature"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, signatureReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("clientBaseUrl"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, baseUrlReader); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signdocs", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}
}

func TestSignWrongPassword(t *testing.T) {
	fileName := "../test/gosign-test-document.pdf"
	certificateName := "../test/47773901104.pfx"

	file, err := io.ReadFile(fileName)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	certificate, err := io.ReadFile(certificateName)
	if err != nil {
		t.Errorf("failed to read certificate")
		t.FailNow()
	}

	fileNameReader := strings.NewReader("gosign-test-document.pdf")
	fileReader := bytes.NewReader(file)
	certificateReader := bytes.NewReader(certificate)
	passwordReader := strings.NewReader("123pmesp")
	signatureReader := strings.NewReader("{ \"gosign-test-document.pdf\": { \"rect\": { \"x\": 309.609375, \"y\": 900.33333333333337, \"page\": 1 }, \"final\": false } }")
	baseUrlReader := strings.NewReader("localhost/verify")

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormField("fileName"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileNameReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("certificate", "certificate.pfx"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, certificateReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("password"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, passwordReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("signature"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, signatureReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("clientBaseUrl"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, baseUrlReader); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signdocs", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}
}

func TestSignBadSignature(t *testing.T) {
	fileName := "../test/gosign-test-document.pdf"
	certificateName := "../test/47773901104.pfx"

	file, err := io.ReadFile(fileName)
	if err != nil {
		t.Errorf("failed to read file")
		t.FailNow()
	}

	certificate, err := io.ReadFile(certificateName)
	if err != nil {
		t.Errorf("failed to read certificate")
		t.FailNow()
	}

	fileNameReader := strings.NewReader("gosign-test-document.pdf")
	fileReader := bytes.NewReader(file)
	certificateReader := bytes.NewReader(certificate)
	passwordReader := strings.NewReader("123pmesp456")
	signatureReader := strings.NewReader("{ \"gosign-test-document.pdf\": { \"rect\": { \"x\": \"309.609375\", \"y\": 900.33333333333337, \"page\": 1 }, \"final\": false } }")
	baseUrlReader := strings.NewReader("localhost/verify")

	var b bytes.Buffer
	multipartWriter := multipart.NewWriter(&b)
	var fw io.Writer

	if fw, err = multipartWriter.CreateFormField("fileName"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileNameReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("file", "test.pdf"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, fileReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormFile("certificate", "certificate.pfx"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, certificateReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("password"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, passwordReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("signature"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, signatureReader); err != nil {
		t.FailNow()
	}

	if fw, err = multipartWriter.CreateFormField("clientBaseUrl"); err != nil {
		t.FailNow()
	}
	if _, err = io.Copy(fw, baseUrlReader); err != nil {
		t.FailNow()
	}

	multipartWriter.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signdocs", &b)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("status: ", w.Code)
		t.FailNow()
	}
}
