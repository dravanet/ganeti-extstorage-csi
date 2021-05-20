package csiclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store"
)

const (
	csiStoragePath = "/srv/ganeti/ganeti-extstorage-csi"
)

// Common errors
var (
	ErrVolumeNotFound           = errors.New("volume not found in store")
	ErrVolumeExists             = errors.New("volume already exists")
	ErrControllerServiceMissing = errors.New("controller service missing")
)

var volumeCapability = &csi.VolumeCapability{
	AccessType: &csi.VolumeCapability_Block{},
	AccessMode: &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	},
}

// New returns a new ganeti-extstorage interface talkint to CSI
func New(endpoint string, tlsConfig *tls.Config, store store.Store) (iface extstorage.Interface, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var opts grpc.DialOption
	if tlsConfig != nil {
		opts = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		opts = grpc.WithInsecure()
	}

	conn, err := grpc.DialContext(ctx, endpoint, opts)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	ic := csi.NewIdentityClient(conn)
	ident, err := ic.GetPluginInfo(ctx, &csi.GetPluginInfoRequest{})
	if err != nil {
		return
	}

	os.Stderr.WriteString(fmt.Sprintf("CSI Name=%s Version=%s manifest=%+v\n", ident.Name, ident.VendorVersion, ident.Manifest))

	caps, err := ic.GetPluginCapabilities(ctx, &csi.GetPluginCapabilitiesRequest{})
	if err != nil {
		return
	}

	cl := &client{
		conn:  conn,
		store: store,
	}

	var volexpansion bool

	for _, cap := range caps.Capabilities {
		if serv := cap.GetService(); serv != nil {
			if serv.GetType() == csi.PluginCapability_Service_CONTROLLER_SERVICE {
				cl.controllerService = true
			}
			if serv.GetType() == csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS {
				err = errors.New("CSI reported VOLUME_ACCESSIBILITY_CONSTRAINTS capability, which is not supported")
				return
			}
		} else if volexp := cap.GetVolumeExpansion(); volexp != nil {
			switch volexp.GetType() {
			case csi.PluginCapability_VolumeExpansion_ONLINE, csi.PluginCapability_VolumeExpansion_OFFLINE:
				volexpansion = true
			}
		}
	}

	if !volexpansion {
		err = errors.New("CSI does not support volume expansion")
		return
	}

	if cl.controllerService {
		var controllerCaps *csi.ControllerGetCapabilitiesResponse

		controller := csi.NewControllerClient(conn)
		controllerCaps, err = controller.ControllerGetCapabilities(ctx, &csi.ControllerGetCapabilitiesRequest{})
		if err != nil {
			return
		}

		for _, cap := range controllerCaps.Capabilities {
			switch cap.GetRpc().GetType() {
			case csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME:
				cl.controllerPublish = true
			}
		}
	}

	return cl, nil
}

func (c *client) Close(ctx context.Context) error {
	return c.conn.Close()
}

type client struct {
	conn  *grpc.ClientConn
	store store.Store

	controllerService bool
	controllerPublish bool
}

// CSIVolumePath returns the target path for a volume
func volumePath(vol *extstorage.VolumeInfo) string {
	return path.Join(csiStoragePath, vol.UUID)
}

func devicePath(vol *extstorage.VolumeInfo) string {
	return path.Join(volumePath(vol), "device")
}

func volumeStagingPath(vol *extstorage.VolumeInfo) string {
	return path.Join(volumePath(vol), "staging")
}
