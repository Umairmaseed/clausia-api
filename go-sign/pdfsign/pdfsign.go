package pdfsign

// #cgo LDFLAGS: -L/usr/local/lib -lpdfsign -lpodofo -lcrypto -lssl
/*
#include <pdfsign.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"unsafe"

	"github.com/goledgerdev/go-sign/logger"
)

// type Details struct {
// 	PM          PolicemanDetail
// 	Assignments []Assignment
// 	Roles       []Role
// }

// PolicemanDetail details a policeman
// type PolicemanDetail struct {
// 	Re       int    `json:"re"`
// 	Name     string `json:"nome"`
// 	RankCode int    `json:"codigoPosto"`
// 	Rank     string `json:"posto"`
// 	OrgCode  int    `json:"codigoOrganizacao"`
// 	OrgDesc  string `json:"descricaoOrganizacao"`
// }

// // Assignment are the functions that a PM has inside the organization
// // this name must change, couldn't find any better
// type Assignment struct {
// 	AssignmentDesc string `json:"AtribDesc"`
// 	AssignmentCode int    `json:"AtribCod"`
// }

// // Role ...d
// type Role struct {
// 	Code int    `json:"FunCod"`
// 	Desc string `json:"DescFuncao"`
// }

// WriteSignature takes an pdf and keys and sign it
func WriteSignature(pubKey, privKey []byte, stampPath, keyPassword string, pdfs ...PDFInput) []Result {
	result := make([]Result, len(pdfs))
	wg := &sync.WaitGroup{}

	for i, pdfFile := range pdfs {
		wg.Add(1)
		fmt.Println(pdfFile)

		go func(routNumber int, file PDFInput) {
			pdf, err := io.ReadAll(file.Pdf)
			if err != nil {
				result[routNumber] = Result{
					FailError: err.Error(),
				}
				return
			}

			// clean those CString allocated
			defer func() {
				wg.Done()
			}()

			qrData := QRData{url: file.ClientURL, id: file.Key, qr: file.QrCode}

			b, err := sign(pdf, pubKey, privKey, keyPassword, stampPath, qrData, file.Param, false)
			if err != nil {
				result[routNumber] = Result{FailError: err.Error()}
				return
			}

			if file.Param.Final {
				certsPath := path.Join("config", "finalCerts")
				if flag.Lookup("test.v") != nil {
					certsPath = path.Join("..", certsPath)
				}

				publicKey, puberr := os.ReadFile(path.Join(certsPath, "certificate.pem"))
				privateKey, priverr := os.ReadFile(path.Join(certsPath, "private-key.pem"))
				if puberr != nil || priverr != nil {
					errStr := "Could not find certificate.pem/private-key.pem pair under config/finalCerts"
					logger.Logger().Sugar().Error("Could not finde certificate.pem/private-key.pem pair under config/finalCerts")
					result[routNumber] = Result{FailError: errStr}
				}

				b, err = sign(
					b, publicKey, privateKey, "", stampPath,
					qrData,
					file.Param, true,
				)
				if err != nil {
					result[routNumber] = Result{FailError: err.Error()}
					return
				}
			}

			result[routNumber] = Result{
				File:       bytes.NewReader(b),
				Filename:   file.Filename,
				SaltedHash: file.SaltedHash,
				QrCode:     file.QrCode,
				Key:        file.Key,
			}

		}(i, pdfFile)
	}
	wg.Wait()
	return result
}

func sign(pdf, pubKey, privKey []byte, pwd, stampPath string, qrdata QRData, param SignatureParam, last bool) ([]byte, error) {
	length := len(pdf)
	b := make([]byte, len(pdf)+300000)
	bufferpointer := (*C.char)(unsafe.Pointer(&b[0]))

	stamp := C.CString(stampPath)

	qrCodeStruct := C.QrCode{
		qrCode: (*C.char)(unsafe.Pointer(&qrdata.qr[0])),
		len:    C.int(len(qrdata.qr)),
		id:     C.CString(qrdata.id),
		url:    C.CString(qrdata.url),
	}
	pdfStr := C.CString(string(pdf))
	signStr := C.CString("Assinatura")
	locStr := C.CString("Brasília")
	creatorStr := C.CString("Criador")
	pubKeyStr := C.CString(string(pubKey))
	privKeyStr := C.CString(string(privKey))
	keyPasswordStr := C.CString(pwd)

	final := C.int(0)
	if last {
		final = C.int(1)
	}

	var detailsdata C.DetailsInfo

	detailsdata.Name = C.CString("Mock name")
	detailsdata.Re = C.CString("")
	detailsdata.Rank = C.CString("Mock rank")

	ret := C.writeSignature(
		pdfStr,
		C.int(length),
		C.SignatureParams{
			signStr,
			locStr,
			creatorStr,
			C.int(param.Rect.X),
			C.int(param.Rect.Y),
			C.int(param.Rect.Page),
			final,
		},
		stamp,
		pubKeyStr,
		privKeyStr,
		keyPasswordStr,
		qrCodeStruct,
		detailsdata,
		bufferpointer,
	)
	if ret == -1 {
		logger.Logger().Sugar().Error("Could not write signature")
		return nil, errors.New("could not write signature")
	}
	free := func() {
		C.free(unsafe.Pointer(pdfStr))
		C.free(unsafe.Pointer(signStr))
		C.free(unsafe.Pointer(locStr))
		C.free(unsafe.Pointer(creatorStr))
		C.free(unsafe.Pointer(pubKeyStr))
		C.free(unsafe.Pointer(privKeyStr))
		C.free(unsafe.Pointer(keyPasswordStr))

		C.free(unsafe.Pointer(detailsdata.Name))
		C.free(unsafe.Pointer(detailsdata.Re))
		C.free(unsafe.Pointer(detailsdata.Rank))
	}
	free()
	b = b[:ret]

	isNull := true
	for _, elem := range b {
		if elem != 0 {
			isNull = false
		}
	}

	if isNull {
		return nil, errors.New("empty document returned by lib go-sign")
	}

	return b, nil
}

type VerifyResult struct {
	Filename string
	Valid    bool
	Key      string
	Error    string
}

func VerifyDocument(pdfs ...PDFInput) []VerifyResult {
	result := make([]VerifyResult, len(pdfs))
	wg := &sync.WaitGroup{}
	for i, pdfFile := range pdfs {
		wg.Add(1)

		go func(routNumber int, file PDFInput) {
			pdf, err := io.ReadAll(file.Pdf)
			result[routNumber] = VerifyResult{
				Valid:    false,
				Filename: file.Filename,
			}
			if err != nil {
				result[routNumber].Error = "error: failed to read PDF"
				log.Println("error: failed to read PDF")
				wg.Done()
				return
			}

			pdfStr := C.CString(string(pdf))
			pwdStr := C.CString(file.Password)

			ret := C.verifySignature(pdfStr, C.int(len(pdf)), pwdStr)

			key, err := RetrieveKey(pdf)
			if err != nil {
				result[routNumber].Error = "error: couldn't retrieve key from PDF"
				log.Println("error: couldn't retrieve key from PDF")
				wg.Done()
				return
			}

			C.free(unsafe.Pointer(pdfStr))
			C.free(unsafe.Pointer(pwdStr))
			if ret == 1 {
				result[routNumber].Valid = true
				result[routNumber].Key = key
				result[routNumber].Error = ""
			}
			wg.Done()
		}(i, pdfFile)

	}
	wg.Wait()

	return result
}

func GetSignatures(pdf []byte) (string, string, error) {
	length := len(pdf)
	cLength := C.int(length)

	pdfStr := C.CString(string(pdf))

	firstSig := make([]byte, 100)
	firstSigPointer := (*C.char)(unsafe.Pointer(&firstSig[0]))

	lastSig := make([]byte, 100)
	lastSigPointer := (*C.char)(unsafe.Pointer(&lastSig[0]))

	retFirst := C.getFirstSignature(pdfStr, cLength, firstSigPointer)
	retLast := C.getLastSignature(pdfStr, cLength, lastSigPointer)

	if retFirst <= 0 {
		logger.Logger().Sugar().Error("Could not retrieve first signature")
		return "", "", errors.New("could not retrieve first signature")
	}

	if retLast <= 0 {
		logger.Logger().Sugar().Error("Could not retrieve last signature")
		return "", "", errors.New("could not retrieve last signature")
	}

	firstSig = firstSig[:retFirst]
	lastSig = lastSig[:retLast]

	return fmt.Sprintf("%x", firstSig), fmt.Sprintf("%x", lastSig), nil
}

func RetrieveKey(pdf []byte) (string, error) {
	pdfStr := C.CString(string(pdf))

	b := make([]byte, 50)

	keybuffer := (*C.char)(unsafe.Pointer(&b[0]))
	keyLen := C.getDocID(pdfStr, C.int(len(pdf)), keybuffer)
	if keyLen == -1 || keyLen == 0 {
		return "", errors.New("could not retrieve document key")
	}

	keyBytes := b[:keyLen]
	key := string(keyBytes)

	fmt.Println("@key: ", key)

	return key, nil
}
