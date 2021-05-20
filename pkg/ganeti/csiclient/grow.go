package csiclient

import (
	"context"
	"os"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

func (c *client) Grow(ctx context.Context, cfg *extstorage.VolumeInfo) error {
	if !c.controllerService {
		return ErrControllerServiceMissing
	}

	vol, err := c.store.Get(ctx, cfg.UUID)
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

	node := csi.NewNodeClient(c.conn)

	nodeCaps, err := node.NodeGetCapabilities(ctx, &csi.NodeGetCapabilitiesRequest{})
	if err != nil {
		return err
	}

	var nodeStage bool
	var nodeExpand bool
	for _, cap := range nodeCaps.Capabilities {
		switch cap.GetRpc().GetType() {
		case csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME:
			nodeStage = true
		case csi.NodeServiceCapability_RPC_EXPAND_VOLUME:
			nodeExpand = true
		}
	}

	// nothing to be done
	if !nodeStage || !nodeExpand {
		return nil
	}

	// Check if volume has already been attached
	volumePath := devicePath(cfg)
	if _, err := os.Stat(volumePath); err != nil {
		return nil
	}

	_, err = node.NodeExpandVolume(ctx, &csi.NodeExpandVolumeRequest{
		VolumeId:          vol.VolumeId,
		VolumePath:        volumePath,
		StagingTargetPath: volumeStagingPath(cfg),
		VolumeCapability:  volumeCapability,
	})

	return err
}
