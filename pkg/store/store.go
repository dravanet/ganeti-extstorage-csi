package store

import (
	"context"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
)

// Store provides a Store where the plugin will store metadata from CSI
type Store interface {
	Add(ctx context.Context, name string, vol *csi.Volume) error
	Get(ctx context.Context, name string) (*csi.Volume, error)
	Remove(ctx context.Context, name string) error
	Close(ctx context.Context) error
}
