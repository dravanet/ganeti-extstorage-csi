package extstorage

import (
	"log"

	"github.com/codingconcepts/env"
)

// VolumeInfo represents Volume information passed by Ganeti
// See https://docs.ganeti.org/docs/ganeti/2.16/html/man-ganeti-extstorage-interface.html#common-environment
type VolumeInfo struct {
	// The name of the volume. This is unique for Ganeti and it uses it to refer to a specific volume inside the external storage.
	// Its format is UUID.ext.diskX where UUID is produced by Ganeti and is unique inside the Ganeti context. X is the number of the disk count.
	Name string `env:"VOL_NAME" required:"true"`
	// Available only to the create and grow scripts. The volume’s size in mebibytes.
	Size int64 `env:"VOL_SIZE"`
	// Available only to the grow script. It declares the new size of the volume after grow (in mebibytes).
	// To find the amount of grow, the scipt should calculate the number VOL_NEW_SIZE - VOL_SIZE.
	NewSize int64 `env:"VOL_NEW_SIZE"`
	// Available only to the setinfo script. A string containing metadata to be associated with the volume.
	// Currently, Ganeti sets this value to originstname+X where X is the instance’s name.
	MetaData string `env:"VOL_METADATA"`
	// The human-readable name of the Disk config object (optional).
	Cname string `env:"VOL_CNAME"`
	// The uuid of the Disk config object.
	UUID string `env:"VOL_UUID"`
	// The name of the volume’s snapshot.
	SnapshotName string `env:"VOL_SNAPSHOT_NAME"`
	// The size of the volume’s snapshot.
	SnapshotSize int64 `env:"VOL_SNAPSHOT_SIZE"`
	// Whether the volume will be opened for exclusive access or not. This will be False (denoting shared access) during migration.
	OpenExclusive bool `env:"VOL_OPEN_EXCLUSIVE"`
}

// ParseVolumeInfo returns VolumeInfo parsed from environment
func ParseVolumeInfo() *VolumeInfo {
	c := &VolumeInfo{}

	if err := env.Set(c); err != nil {
		log.Fatal(err)
	}

	return c
}
