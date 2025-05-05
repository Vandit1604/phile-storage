package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Configurations for Etcd
const (
	dialTimeout = 5 * time.Second
)

// EtcdClient wraps the etcd client
type EtcdClient struct {
	client *clientv3.Client
}

// Creates a New EtcdClient
func NewEtcdClient(endpoints []string) (*EtcdClient, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdClient{client: client}, nil
}

// Put sets a KV pair in etcd
func (e *EtcdClient) Put(ctx context.Context, key, value string, ttl int64) error {
	lease, err := e.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	_, err = e.client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	return err
}

// Get retrieves a KV pair from etcd
func (e *EtcdClient) Get(ctx context.Context, key string) (string, error) {
	etcdResp, err := e.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(etcdResp.Kvs) == 0 {
		return "", nil
	}

	return string(etcdResp.Kvs[0].Value), nil
}

// Delete removes a KV pair from etcd
func (e *EtcdClient) Delete(ctx context.Context, key string) error {
	_, err := e.client.Delete(ctx, key)
	return err
}

// Shuts down Etcd client
func (e *EtcdClient) Close() error {
	err := e.client.Close()
	return err
}
