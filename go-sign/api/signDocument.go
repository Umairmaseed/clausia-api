package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/go-sign/logger"
	"github.com/goledgerdev/go-sign/pdfsign"
	qrcode "github.com/skip2/go-qrcode"
	"software.sslmate.com/src/go-pkcs12"
)

type signForm struct {
	FileName      string                `form:"fileName" binding:"required"`
	File          *multipart.FileHeader `form:"file" binding:"required"`
	Certificate   *multipart.FileHeader `form:"certificate" binding:"required"`
	Password      string                `form:"password" biding:"required"`
	Signature     string                `form:"signature" binding:"required"`
	LedgerKey     string                `form:"ledgerKey"`
	ClientBaseUrl string                `form:"clientBaseUrl" binding:"required"`
}

type SignaturesObj map[string]pdfsign.SignatureParam

type fileResponse struct {
	Name          string `json:"name"`
	File          []byte `json:"file"`
	PubKey        string `json:"pubKey"`
	QRCode        string `json:"qrCode"`
	SaltedHash    string `json:"saltedHash"`
	LastSignature string `json:"lastSignature"`
	Key           string `json:"key"`
}

func SignDocument(c *gin.Context) {
	var form signForm
	log := logger.Logger().Sugar()
	if err := c.Bind(&form); err != nil {
		log.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var final SignaturesObj
	err := json.Unmarshal([]byte(form.Signature), &final)
	if err != nil {
		log.Error(err)
		c.String(http.StatusBadRequest, "Could not decode signature info from body")
		return
	}

	certificateOpen, err := form.Certificate.Open()
	if err != nil {
		log.Error("Could not open file from form", err)
		c.String(400, "missing certificate file")
		return
	}
	certificate, _ := io.ReadAll(certificateOpen)
	priv, cert, _, err := pkcs12.DecodeChain(certificate, form.Password)
	if err != nil {
		log.Error("Could not open certificate, err: ", err)
		c.String(400, "could not open certificate: %s", err.Error())
		return
	}

	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		log.Error("Could not bind key")
		c.String(http.StatusInternalServerError, "criptographic error")
		return
	}

	block := pem.Block{Bytes: cert.Raw, Type: "CERTIFICATE"}

	x509Encoded, _ := x509.MarshalECPrivateKey(ecdsaPriv)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	pubKey := pem.EncodeToMemory(&block)

	fileName := form.FileName
	// fileBytes := form.File

	openedFile, err := form.File.Open()
	if err != nil {
		log.Error("Could not open file from form", err)
		c.String(400, "missing pdf file")
		return
	}

	pdfByte, err := io.ReadAll(openedFile)
	if err != nil {
		log.Error("failed to read pdf bytes: ", err)
		c.String(http.StatusInternalServerError, "falha na abertura do PDF")
		return
	}

	ledgerKey := ""
	ledgerKey, err = pdfsign.RetrieveKey(pdfByte)
	if err != nil {
		log.Infof("could not retrieve ledgerKey form PDF, creating new asset: %s", err.Error())
	}

	if ledgerKey == "" {
		log.Info("pdf has no ledgerKey defined, getting from blockchain")

		ledgerKey = form.LedgerKey

		if ledgerKey == "" {
			log.Error("Could not add ledger key")
			c.String(http.StatusBadRequest, "missing ledger key")
			return
		}

	}

	log.Info("ledgerKey is: ", ledgerKey)

	var splitKey string

	if strings.Contains(ledgerKey, ":") {
		s := strings.Split(ledgerKey, ":")
		splitKey = s[1]
	} else {
		splitKey = ledgerKey
	}

	qr, err := qrcode.Encode(form.ClientBaseUrl+splitKey, qrcode.Medium, 256)
	if err != nil {
		log.Error("failed to generate qr code for document hash: ", err)
		c.String(http.StatusInternalServerError, "falha na geração do QR Code")
		return
	}

	pdf := pdfsign.PDFInput{
		Filename:  fileName,
		Pdf:       bytes.NewReader(pdfByte),
		Param:     final[fileName],
		QrCode:    qr,
		ClientURL: form.ClientBaseUrl,
		Key:       splitKey,
	}

	signResults := pdfsign.WriteSignature(pubKey, pemEncoded, "./fixtures/stamps/goledger-icon.png", form.Password, pdf)
	for _, res := range signResults {
		if res.FailError != "" {
			log.Error("failed to sign document: ", res.FailError)
			c.String(http.StatusInternalServerError, "falha na assinatura do PDF")
			return
		}
	}

	var response fileResponse

	res := signResults[0]
	if res.File != nil {
		file, _ := io.ReadAll(res.File)
		_, lastSig, err := pdfsign.GetSignatures(file)
		if err != nil {
			log.Errorf("failed to get signatures for the file", res.FailError)
			c.String(http.StatusInternalServerError, "Não foi possível obter as assinaturas do arquivo")
			return
		}
		response = fileResponse{
			Name:          strings.TrimSuffix(res.Filename, filepath.Ext(res.Filename)),
			File:          file,
			PubKey:        string(pubKey),
			QRCode:        base64.StdEncoding.EncodeToString(res.QrCode),
			Key:           res.Key,
			SaltedHash:    res.SaltedHash,
			LastSignature: lastSig,
		}
	} else {
		log.Errorf("returned file is empty")
		c.String(http.StatusInternalServerError, "returned file is empty")
		return
	}

	c.JSON(200, response)
}
