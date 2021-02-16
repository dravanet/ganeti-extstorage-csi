// Package file provides a simple file-based implementation
// for store. This should not be used in production.
package file

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store"
)

// New returns a file-based Store
func New(storeBase string) (store.Store, error) {
	if err := os.MkdirAll(storeBase, 0o750); err != nil {
		return nil, err
	}

	return &file{
		base: storeBase,
	}, nil
}

type file struct {
	base string
}

func (s *file) path(name string) string {
	return path.Join(s.base, name)
}

func (s *file) Add(ctx context.Context, name string, vol *csi.Volume) error {
	metadatapath := s.path(name)

	if _, err := os.Stat(metadatapath); err == nil || !os.IsNotExist(err) {
		return err
	}

	data, err := json.Marshal(vol)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(metadatapath, data, 0o640)
}

func (s *file) Get(ctx context.Context, name string) (*csi.Volume, error) {
	metadatapath := s.path(name)

	data, err := ioutil.ReadFile(metadatapath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var vol csi.Volume
	if err = json.Unmarshal(data, &vol); err != nil {
		return nil, err
	}

	return &vol, nil
}

func (s *file) Remove(ctx context.Context, name string) error {
	metadatapath := s.path(name)

	return os.Remove(metadatapath)
}

func (s *file) Close(ctx context.Context) error {
	return nil
}
