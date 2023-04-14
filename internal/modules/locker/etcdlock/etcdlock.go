package etcdlock

import (
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"lucky/internal/config"
	"lucky/internal/modules/locker"
	"sync"
)

type EtcdLockImpl struct {
	client *concurrency.Session
	data   sync.Map
}

func New(config config.EtcdConfig) (locker.Locker, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: config.Endpoints,
	})
	if err != nil {
		return nil, err
	}
	// Test client with ping
	_, err = client.Status(context.Background(), config.Endpoints[0])
	//fmt.Println("Etcd client Ping response:", resp)
	if err != nil {
		return nil, err
	}
	// create a sessions to acquire a lock
	s, _ := concurrency.NewSession(client)
	//defer s.Close()
	if err != nil {
		return nil, err
	}
	return &EtcdLockImpl{
		client: s,
	}, nil
}

func (m *EtcdLockImpl) Lock(key string) error {
	// Obtain a new mutex by using the same name for all instances wanting the same lock.
	mutex, _ := m.data.LoadOrStore(key, concurrency.NewMutex(m.client, key))
	ctx := context.Background()
	if err := mutex.(*concurrency.Mutex).Lock(ctx); err != nil {
		return err
	}
	return nil
}

func (m *EtcdLockImpl) Unlock(key string) error {
	// Obtain a new mutex by using the same name for all instances wanting the same lock.
	mutex, exists := m.data.Load(key)
	if !exists {
		return errors.New("key not found")
	}
	ctx := context.Background()
	if err := mutex.(*concurrency.Mutex).Unlock(ctx); err != nil {
		return err
	}

	return nil

}