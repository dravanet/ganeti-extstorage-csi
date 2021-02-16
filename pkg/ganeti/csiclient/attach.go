package csiclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/ganeti/extstorage"
)

func (c *client) Attach(ctx context.Context, cfg *extstorage.VolumeInfo) error {
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

	volPath := volumePath(cfg)
	os.MkdirAll(volPath, 0o755)
	var stagingTargetPath string

	if nodeStage {
		stagingTargetPath = volumeStagingPath(cfg)
		os.MkdirAll(stagingTargetPath, 0o750)

		_, err = node.NodeStageVolume(ctx, &csi.NodeStageVolumeRequest{
			VolumeId:          vol.VolumeId,
			VolumeContext:     vol.VolumeContext,
			StagingTargetPath: stagingTargetPath,
			VolumeCapability:  volumeCapability,
		})

		if err != nil {
			return err
		}
	}

	targetPath := devicePath(cfg)
	_, err = node.NodePublishVolume(ctx, &csi.NodePublishVolumeRequest{
		VolumeId:          vol.VolumeId,
		VolumeContext:     vol.VolumeContext,
		StagingTargetPath: stagingTargetPath,
		TargetPath:        targetPath,
		VolumeCapability:  volumeCapability,
	})
	if err != nil {
		return err
	}

	rpath, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		return err
	}
	fmt.Println(rpath)

	return nil
}
