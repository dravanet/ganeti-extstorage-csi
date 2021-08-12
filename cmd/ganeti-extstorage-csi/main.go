package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"time"

	"github.com/namsral/flag"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/csiclient"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store/etcd"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store/file"
)

var (
	// CSI variables
	csiEndpoint       = flag.String("csi-endpoint", "unix:///csi/csi.sock", "CSI endpoint to connect to")
	csiTlsCert        = flag.String("csi-tls-cert", "", "CSI TLS Client Certificate")
	csiTlsKey         = flag.String("csi-tls-key", "", "CSI TLS Client Private key")
	csiTlsCA          = flag.String("csi-tls-ca", "", "CSI TLS Certificate Authority")
	operation         = flag.String("operation", "", "Operation to perform: create|attach|detach|remove|grow|setinfo|verify")
	etcdStoreEndpoint = flag.String("etcd-store-endpoint", "localhost:2379", "Etcd endpoint for etcd store")
	etcdTlsCert       = flag.String("etcd-tls-cert", "", "Etcd TLS Client Certificate")
	etcdTlsKey        = flag.String("etcd-tls-key", "", "Etcd TLS Client Private key")
	etcdTlsCA         = flag.String("etcd-tls-ca", "", "Etcd TLS Certificate Authority")
	fileStoreBase     = flag.String("file-store-base", "", "File store base directory, for development")
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	var st store.Store
	var err error
	var tlsConfig *tls.Config

	flag.Parse()

	if *fileStoreBase != "" {
		st, err = file.New(*fileStoreBase)
	} else {
		tlsConfig, err = prepareTlsConfig(*etcdTlsCert, *etcdTlsKey, *etcdTlsCA)
		if err != nil {
			log.Fatalf("Error preparing tls configuration for etcd: %+v", err)
		}
		st, err = etcd.New(*etcdStoreEndpoint, tlsConfig)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close(ctx)

	tlsConfig, err = prepareTlsConfig(*csiTlsCert, *csiTlsKey, *csiTlsCA)
	if err != nil {
		log.Fatalf("Error preparing tls configuration for csi: %+v", err)
	}

	volConfig := extstorage.ParseVolumeInfo()

	client, err := csiclient.New(*csiEndpoint, tlsConfig, st)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)

	switch *operation {
	case "create":
		err = client.Create(ctx, volConfig)
	case "attach":
		err = client.Attach(ctx, volConfig)
	case "detach":
		err = client.Detach(ctx, volConfig)
	case "remove":
		err = client.Remove(ctx, volConfig)
	case "grow":
		err = client.Grow(ctx, volConfig)
	case "setinfo":
		err = client.Setinfo(ctx, volConfig)
	case "verify":
		err = client.Verify(ctx, volConfig)
	default:
		err = errors.New("Invalid command")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func prepareTlsConfig(certFile, keyFile, caFile string) (tlsConfig *tls.Config, err error) {
	if certFile != "" && keyFile != "" {
		tlsConfig = &tls.Config{}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

		if caFile != "" {
			roots := x509.NewCertPool()
			cacerts, err := os.ReadFile(caFile)
			if err != nil {
				return nil, err
			}
			roots.AppendCertsFromPEM(cacerts)

			// Perform server validation, only check CA trust model
			tlsConfig.InsecureSkipVerify = true
			tlsConfig.VerifyConnection = func(cs tls.ConnectionState) error {
				opts := x509.VerifyOptions{
					Intermediates: x509.NewCertPool(),
					Roots:         roots,
					KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
				}
				for _, cert := range cs.PeerCertificates[1:] {
					opts.Intermediates.AddCert(cert)
				}
				_, err := cs.PeerCertificates[0].Verify(opts)
				return err
			}
		}
	}

	return
}
