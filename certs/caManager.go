package certs

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"sync"

	"github.com/google/logger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"go.uber.org/zap"
	"software.sslmate.com/src/go-pkcs12"
)

var (
	ca   *CAMngr
	once sync.Once
)

type CAMngr struct {
	sdk           *fabsdk.FabricSDK
	logger        *zap.SugaredLogger
	msp           *msp.Client
	sdkConfigPath string
	caURL         string
}

func InitCAMngr(sdkConfigPath, caURL string) (*CAMngr, error) {
	once.Do(func() {
		sdk, err := fabsdk.New(config.FromFile(sdkConfigPath))
		if err != nil {
			logger.Error("failed to create new sdk", err)
			zap.S().Errorf("Failed to create new SDK: %v", err)
			return
		}

		client, err := msp.New(sdk.Context())
		if err != nil {
			logger.Error("failed to create new msp client", err)
			zap.S().Errorf("Failed to create new MSP client: %v", err)
			return
		}

		ca = &CAMngr{
			sdk:           sdk,
			logger:        zap.S(),
			msp:           client,
			sdkConfigPath: sdkConfigPath,
			caURL:         caURL,
		}
	})
	if ca == nil {
		return nil, errors.New("failed to initialize CA manager")
	}
	return ca, nil
}

func (c *CAMngr) CreateIdentity(username, commonName, password string) ([]byte, error) {
	_, err := c.msp.CreateIdentity(&msp.IdentityRequest{
		Secret: password,
		ID:     username,
	})
	if err != nil {
		c.logger.Errorf("Could not create identity [%s]: %s", username, err)
		return nil, err
	}

	err = c.msp.Enroll(username, msp.WithSecret(password), msp.WithCSR(&msp.CSRInfo{CN: commonName}))
	if err != nil {
		c.logger.Errorf("Could not enroll identity [%s]: %s", username, err)
		return nil, err
	}

	return c.getPFX(username, password)
}

func (c *CAMngr) getPFX(username, certPwd string) ([]byte, error) {
	private, publicCert, err := c.getKeyPair(username)
	if err != nil {
		c.logger.Errorf("Could not get key pair for identity [%s]: %s", username, err)
		return nil, err
	}

	data, err := pkcs12.Encode(rand.Reader, private, publicCert, nil, certPwd)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	return data, nil
}

func (c *CAMngr) getKeyPair(username string) (interface{}, *x509.Certificate, error) {
	si, err := c.msp.GetSigningIdentity(username)
	if err != nil {
		c.logger.Errorf("Could not get signing identity for user [%s]: %s", username, err)
		return nil, nil, err
	}

	priv, err := si.PrivateKey().Bytes()
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(priv)
	if block == nil {
		return nil, nil, errors.New("could not decode the private key block")
	}

	private, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	block, _ = pem.Decode(si.EnrollmentCertificate())
	if block == nil {
		return nil, nil, errors.New("could not decode the certificate block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return private, cert, nil
}
