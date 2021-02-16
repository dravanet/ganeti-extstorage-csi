package csiclient

import (
	"context"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

func (c *client) Grow(ctx context.Context, cfg *extstorage.VolumeInfo) error {
	vol, err := c.store.Get(ctx, cfg.Name)
	if err != nil {
		return err
	}

	if vol == nil {
		return ErrVolumeNotFound
	}

	cont := csi.NewControllerClient(c.conn)

	_, err = cont.ControllerExpandVolume(ctx, &csi.ControllerExpandVolumeRequest{
		VolumeId: vol.VolumeId,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: cfg.NewSize * mebibytes,
			LimitBytes:    cfg.NewSize * mebibytes,
		},
		VolumeCapability: volumeCapability,
	})
	if err != nil {
		return err
	}

	return nil
}
