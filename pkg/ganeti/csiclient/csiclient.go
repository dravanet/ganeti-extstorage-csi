package csiclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store"
)

const (
	csiStoragePath = "/srv/ganeti/ganeti-extstorage-csi"
)

// Common errors
var (
	ErrVolumeNotFound = errors.New("Volume not found in store")
	ErrVolumeExists   = errors.New("Volume already exists")
)

var volumeCapability = &csi.VolumeCapability{
	AccessType: &csi.VolumeCapability_Block{},
	AccessMode: &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	},
}

// New returns a new ganeti-extstorage interface talkint to CSI
func New(endpoint string, store store.Store) (iface extstorage.Interface, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint, grpc.WithInsecure())
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

	var volexpansion bool

	for _, cap := range caps.Capabilities {
		if serv := cap.GetService(); serv != nil {
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

	controller := csi.NewControllerClient(conn)
	controllerCaps, err := controller.ControllerGetCapabilities(ctx, &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		return
	}

	cl := &client{
		conn:  conn,
		store: store,
	}

	for _, cap := range controllerCaps.Capabilities {
		switch cap.GetRpc().GetType() {
		case csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME:
			cl.controllerPublish = true
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

	controllerPublish bool
}

// CSIVolumePath returns the target path for a volume
func volumePath(vol *extstorage.VolumeInfo) string {
	return path.Join(csiStoragePath, vol.Name)
}

func devicePath(vol *extstorage.VolumeInfo) string {
	return path.Join(volumePath(vol), "device")
}

func volumeStagingPath(vol *extstorage.VolumeInfo) string {
	return path.Join(volumePath(vol), "staging")
}
