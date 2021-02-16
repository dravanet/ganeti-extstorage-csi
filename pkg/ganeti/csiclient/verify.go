package csiclient

import (
	"context"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

func (c *client) Verify(ctx context.Context, cfg *extstorage.VolumeInfo) error {
	vol, err := c.store.Get(ctx, cfg.Name)
	if err != nil {
		return err
	}

	if vol == nil {
		return ErrVolumeNotFound
	}

	return nil
}
