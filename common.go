package gohttpsserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

const rsaBits = 2048
const isCA = false
const validYears = 1

// end of ASN.1 time
//var endOfTime = time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC)

func NewSelfSignedCertificate() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now().Add(-5 * time.Minute).UTC()
	notAfter := notBefore.AddDate(validYears, 0, 0).UTC()

	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			Organization: []string{"Example Inc"},
			// does not seem to be required, but makes it more similar to "real" keys
			CommonName: "localhost",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := []string{"localhost", "example.com", "127.0.0.1"}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
}

func ServeWithCertAndKey(addr string, certificate tls.Certificate, handler http.Handler) error {
	if addr == "" {
		addr = ":https"
	}
	server := &http.Server{Addr: addr, Handler: handler}

	tlsConfig := tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{certificate}

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(conn, &tlsConfig)
	return server.Serve(tlsListener)
}

func ServeWithGeneratedCert(addr string, handler http.Handler) error {
	certificate, err := NewSelfSignedCertificate()
	if err != nil {
		return err
	}
	return ServeWithCertAndKey(addr, certificate, handler)
}
