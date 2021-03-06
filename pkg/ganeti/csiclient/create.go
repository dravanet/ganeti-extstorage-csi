package csiclient

import (
	"context"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

const mebibytes = 1 << 20

func (c *client) Create(ctx context.Context, cfg *extstorage.VolumeInfo) error {
	if !c.controllerService {
		return ErrControllerServiceMissing
	}

	vol, err := c.store.Get(ctx, cfg.UUID)
	if err != nil {
		return err
	}

	if vol != nil {
		return ErrVolumeExists
	}

	cont := csi.NewControllerClient(c.conn)

	resp, err := cont.CreateVolume(ctx, &csi.CreateVolumeRequest{
		Name: cfg.UUID,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: cfg.Size * mebibytes,
			LimitBytes:    cfg.Size * mebibytes,
		},
		VolumeCapabilities: []*csi.VolumeCapability{volumeCapability},
	})
	if err != nil {
		return err
	}

	return c.store.Add(ctx, cfg.UUID, resp.Volume)
}
