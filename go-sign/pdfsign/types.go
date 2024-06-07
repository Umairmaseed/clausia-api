package pdfsign

import "io"

type Result struct {
	File       io.Reader
	Filename   string
	FailError  string
	SaltedHash string
	QrCode     []byte
	Key        string
}

type SignatureParam struct {
	Final bool `json:"final"`
	Rect  struct {
		X    float64 `json:"x"`
		Y    float64 `json:"y"`
		Page int     `json:"page"`
	} `json:"rect"`
}

type PDFInput struct {
	Filename   string
	Pdf        io.Reader
	Param      SignatureParam
	Password   string
	QrCode     []byte
	ClientURL  string
	SaltedHash string
	Key        string
}

type QRData struct {
	url string
	id  string
	qr  []byte
}
