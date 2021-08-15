package generators

// import (
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"crypto/x509"
// 	"crypto/x509/pkix"
// 	"encoding/pem"
// 	"math/big"
// 	"net"
// 	"time"
// )

// var (
// 	keyBitSize = 4096
// )

// const (
// 	// clientCertType represents client certificates
// 	clientCertType = "client"
// 	// caCertType represents a CA certificate
// 	caCertType = "ca"
// )

// func getCertTemplate(certificateType string) x509.Certificate {
// 	cert := x509.Certificate{
// 		SerialNumber: big.NewInt(1),
// 		Subject: pkix.Name{
// 			Organization: []string{"Pachyderm, Inc."},
// 			Country:      []string{"US"},
// 		},
// 		NotBefore:             time.Now().UTC(),
// 		NotAfter:              time.Now().Add(time.Hour * 24 * 366).UTC(),
// 		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 	}

// 	if certificateType == caCertType {
// 		cert.KeyUsage |= x509.KeyUsageCertSign
// 		cert.IsCA = true
// 	}

// 	if certificateType == clientCertType {
// 		cert.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
// 		cert.IsCA = false
// 	}

// 	return cert
// }

// // PrivateKeyRSA returns a RSA private key
// func newPrivateKeyRSA(size int) (*rsa.PrivateKey, error) {
// 	return rsa.GenerateKey(rand.Reader, size)
// }

// // EncodePrivateKeyToPEM returns privateKey in PEM format
// func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
// 	return pem.EncodeToMemory(
// 		&pem.Block{
// 			Type:  "RSA PRIVATE KEY",
// 			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
// 		},
// 	)
// }

// // EncodeCertificateToPEM converts a certificate to PEM format
// func encodeCertificateToPEM(cert *x509.Certificate) []byte {
// 	return pem.EncodeToMemory(
// 		&pem.Block{
// 			Type:  "CERTIFICATE",
// 			Bytes: cert.Raw,
// 		},
// 	)
// }

// // ClientCertificate returns certificate, privateKey and error
// // Takes an RSA private key and slice of hosts
// // Returns: certificate, error
// func newClientCertificate(key *rsa.PrivateKey, hosts []string) (*x509.Certificate, error) {

// 	cert := getCertTemplate(clientCertType)

// 	if len(hosts) > 0 {
// 		cert.DNSNames = append(cert.DNSNames, hosts...)
// 	}

// 	certificateDER, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &key.PublicKey, key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return x509.ParseCertificate(certificateDER)
// }
