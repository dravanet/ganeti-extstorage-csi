package csiclient

import (
	"context"
	"os"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

func (c *client) Detach(ctx context.Context, cfg *extstorage.VolumeInfo) error {
	vol, err := c.store.Get(ctx, cfg.Name)
	if err != nil {
		return err
	}

	if vol == nil {
		return ErrVolumeNotFound
	}

	node := csi.NewNodeClient(c.conn)

	nodeCaps, err := node.NodeGetCapabilities(ctx, &csi.NodeGetCapabilitiesRequest{})
	if err != nil {
		return err
	}

	var nodeStage bool
	for _, cap := range nodeCaps.Capabilities {
		if cap.GetRpc().GetType() == csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME {
			nodeStage = true
		}
	}

	_, err = node.NodeUnpublishVolume(ctx, &csi.NodeUnpublishVolumeRequest{
		VolumeId:   vol.VolumeId,
		TargetPath: devicePath(cfg),
	})
	if err != nil {
		return err
	}

	if nodeStage {
		stagingTargetPath := volumeStagingPath(cfg)

		_, err = node.NodeUnstageVolume(ctx, &csi.NodeUnstageVolumeRequest{
			VolumeId:          vol.VolumeId,
			StagingTargetPath: stagingTargetPath,
		})

		if err != nil {
			return err
		}

		os.Remove(stagingTargetPath)
	}

	os.Remove(volumePath(cfg))

	return nil
}
