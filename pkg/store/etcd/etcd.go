// Package etcd provides an etcd based implementation
package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	v3 "go.etcd.io/etcd/api/v3/etcdserverpb"
	"google.golang.org/grpc"

	"github.com/dravanet/ganeti-extstorage-csi/pkg/csi"
	"github.com/dravanet/ganeti-extstorage-csi/pkg/store"
)

const (
	keyPrefix = "volmeta"
)

// New returns an etcd based Store
func New(endpoint string) (store.Store, error) {
	var options []grpc.DialOption
	switch {
	case strings.HasPrefix(endpoint, "http://"):
		endpoint = strings.TrimPrefix(endpoint, "http://")
		options = append(options, grpc.WithInsecure())
	case strings.HasPrefix(endpoint, "https://"):
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}

	conn, err := grpc.Dial(endpoint, options...)
	if err != nil {
		return nil, err
	}

	return &etcd{
		conn: conn,
		kv:   v3.NewKVClient(conn),
	}, nil
}

type etcd struct {
	conn *grpc.ClientConn
	kv   v3.KVClient
}

func (s *etcd) Add(ctx context.Context, name string, vol *csi.Volume) error {
	data, err := json.Marshal(vol)
	if err != nil {
		return err
	}

	key := keyFromVol(name)

	resp, err := s.kv.Txn(ctx, &v3.TxnRequest{
		Compare: []*v3.Compare{
			{
				Key:         key,
				Target:      v3.Compare_CREATE,
				Result:      v3.Compare_EQUAL,
				TargetUnion: &v3.Compare_CreateRevision{CreateRevision: 0},
			},
		},
		Success: []*v3.RequestOp{
			{
				Request: &v3.RequestOp_RequestPut{
					RequestPut: &v3.PutRequest{
						Key:   key,
						Value: data,
					},
				},
			},
		},
	})

	if err != nil {
		return err
	}

	if resp.Succeeded != true {
		return errors.New("already exists")
	}

	return err
}

func (s *etcd) Get(ctx context.Context, name string) (*csi.Volume, error) {
	resp, err := s.kv.Range(ctx, &v3.RangeRequest{
		Key: keyFromVol(name),
	})

	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, nil
	}

	var vol csi.Volume

	if err = json.Unmarshal(resp.Kvs[0].Value, &vol); err != nil {
		return nil, err
	}

	return &vol, nil
}

func (s *etcd) Remove(ctx context.Context, name string) error {
	_, err := s.kv.DeleteRange(ctx, &v3.DeleteRangeRequest{
		Key: keyFromVol(name),
	})

	return err
}

func (s *etcd) Close(ctx context.Context) error {
	return s.conn.Close()
}

func keyFromVol(name string) []byte {
	return []byte(fmt.Sprintf("%s/%s", keyPrefix, name))
}
