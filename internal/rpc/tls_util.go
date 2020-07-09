package rpc

import (
	"fmt"

	"crypto/tls"
	"crypto/x509"

	"github.com/pkg/errors"
)

func trustedCertificateFromFileData(publicCertFileData []byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(publicCertFileData) {
		return nil, errors.WithStack(fmt.Errorf("trustedCertificateFromFileData: failed to append certificates"))
	}
	return pool, nil
}

func certificateFromFileData(publicCertFileData []byte, privateKeyFileData []byte) (*tls.Certificate, error) {
	cert, err := tls.X509KeyPair(publicCertFileData, privateKeyFileData)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &cert, nil
}