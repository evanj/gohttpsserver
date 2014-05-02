package gohttpsserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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

func getRandomSerial() int64 {
	var id int64 = 0
	// do not permit an id of zero
	for id == 0 {
		err := binary.Read(rand.Reader, binary.LittleEndian, &id)
		if err != nil {
			panic("binary.Read failed: " + err.Error())
		}
	}

	// clear the top bit to force it to be positive
	id &= ^(math.MinInt64)
	return id
}

var defaultHosts = []string{"localhost", "127.0.0.1"}

// Based on generate_cert:
// https://code.google.com/p/go/source/browse/src/pkg/crypto/tls/generate_cert.go
func NewSelfSignedCertificate(hosts []string) (tls.Certificate, error) {
	if len(hosts) == 0 {
		hosts = defaultHosts
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now().Add(-5 * time.Minute).UTC()
	notAfter := notBefore.AddDate(validYears, 0, 0).UTC()

	template := x509.Certificate{
		// must be unique to avoid errors when serial/issuer is reused with different keys
		SerialNumber: new(big.Int).SetInt64(getRandomSerial()),
		Subject: pkix.Name{
			Organization: []string{"Example Inc"},
			// does not seem to be required, but makes it more similar to "real" keys
			CommonName: hosts[0],
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

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

func Serve(addr string, certificate tls.Certificate, handler http.Handler) error {
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

func ServeWithNewSelfSigned(addr string, handler http.Handler) error {
	certificate, err := NewSelfSignedCertificate(nil)
	if err != nil {
		return err
	}
	return Serve(addr, certificate, handler)
}

func makeProxyHeaderDirector(originalDirector func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		protocol := "http"
		if r.TLS != nil {
			protocol = "https"
		}
		log.Printf("%s %s://%s%s %s", r.Method, protocol, r.Host, r.URL, r.RemoteAddr)
		// X-Forwarded-For is set by ReverseProxy; X-Forwarded-Proto is not
		r.Header.Set("X-Forwarded-Proto", protocol)

		originalDirector(r)
	}
}

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = makeProxyHeaderDirector(proxy.Director)
	return proxy
}
