package etcd

import (
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

// Shuts down Etcd client
func (e *EtcdClient) Close() error {
	err := e.client.Close()
	return err
}
