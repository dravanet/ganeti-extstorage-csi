package csiclient

import (
	"context"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

func (c *client) Remove(ctx context.Context, cfg *extstorage.VolumeInfo) error {
	vol, err := c.store.Get(ctx, cfg.UUID)
	if err != nil {
		return err
	}

	if vol == nil {
		return ErrVolumeNotFound
	}

	cont := csi.NewControllerClient(c.conn)

	_, err = cont.DeleteVolume(ctx, &csi.DeleteVolumeRequest{
		VolumeId: vol.VolumeId,
	})

	if err != nil {
		return err
	}

	return c.store.Remove(ctx, cfg.UUID)
}
