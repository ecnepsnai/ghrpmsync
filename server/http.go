package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ecnepsnai/web/router"
)

func getPortFromArg(argName string, defaultPort int) (port int) {
	port = defaultPort
	args := os.Args
	if len(args) == 1 {
		return
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == argName {
			if i == len(args)-1 {
				log.Fatal("Argument %s requires a value", arg)
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil {
				log.Fatal("Argument %s requires a numeric value", arg)
			}
			port = p
			return
		}
	}

	return
}

func startHTTPS(server *router.Server) error {
	var pKey crypto.PrivateKey
	var err error
	pKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("crypto error: %s", err.Error())
	}

	pub := pKey.(crypto.Signer).Public()
	serial := big.NewInt(1)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return fmt.Errorf("crypto error: %s", err.Error())
	}
	h := sha1.Sum(publicKeyBytes)

	tpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: ""},
		NotBefore:             time.Unix(0, 0),
		NotAfter:              time.Now().UTC().AddDate(100, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		BasicConstraintsValid: true,
		SubjectKeyId:          h[:],
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, pKey)
	if err != nil {
		return fmt.Errorf("crypto error: %s", err.Error())
	}

	address := fmt.Sprintf("0.0.0.0:%d", getPortFromArg("--https-port", 443))
	l, err := tls.Listen("tcp", address, &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certBytes},
				PrivateKey:  pKey,
			},
		},
	})
	log.PDebug("Listen", map[string]interface{}{
		"address": address,
	})
	if err != nil {
		return fmt.Errorf("listen error: %s", err.Error())
	}
	return server.Serve(l)
}

func startHTTP(server *router.Server) error {
	return server.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", getPortFromArg("--http-port", 80)))
}
