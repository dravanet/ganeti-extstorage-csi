package main

import (
	"context"
	"errors"
	"log"
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
	operation         = flag.String("operation", "", "Operation to perform: create|attach|detach|remove|grow|setinfo|verify")
	etcdStoreEndpoint = flag.String("etcd-store-endpoint", "http://localhost:2379", "Etcd endpoint for etcd store")
	fileStoreBase     = flag.String("file-store-base", "", "File store base directory, for development")
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	var st store.Store
	var err error

	flag.Parse()

	if *fileStoreBase != "" {
		st, err = file.New(*fileStoreBase)
	} else {
		st, err = etcd.New(*etcdStoreEndpoint)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close(ctx)

	volConfig := extstorage.ParseVolumeInfo()

	client, err := csiclient.New(*csiEndpoint, st)
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
