package pool

import (
	"fmt"
	"testing"
	"time"
)

func TestPoolPutGet(t *testing.T) {
	var (
		user     = "root"
		host     = "127.0.0.1"
		port     = 2222
		password = "secret"
	)

	go simpleSSHServer()
	clientPool := NewPool(time.Minute*10, time.Minute*10, 1000, 3)

	client, err := NewSSHClient(user, host, port, SetPassword(password))
	if err != nil {
		t.Errorf("NewSSHClient error:%v", err)
	}

	cacheKey := fmt.Sprintf("%s@%s:%d", user, host, port)

	err = clientPool.Put(cacheKey, client)
	if err != nil {
		t.Errorf("Put error:%v", err)
	}

	_, ok := clientPool.Get(cacheKey)
	if !ok {
		t.Errorf("Test Get failed")
	}
}

func TestPoolGetWithNew(t *testing.T) {
	var (
		user     = "root"
		host     = "127.0.0.1"
		port     = 2222
		password = "secret"
	)

	go simpleSSHServer()
	clientPool := NewPool(time.Minute*10, time.Minute*10, 1000, 3)

	cacheKey := fmt.Sprintf("%s@%s:%d", user, host, port)
	_, err := clientPool.GetWithNew(cacheKey, user, host, port, SetPassword(password))
	if err != nil {
		t.Errorf("GetWithNew error:%v", err)
	}
}

func TestPool_Delete(t *testing.T) {
	var (
		user     = "root"
		host     = "127.0.0.1"
		port     = 2222
		password = "secret"
	)

	go simpleSSHServer()
	clientPool := NewPool(time.Minute*10, time.Minute*10, 1000, 3)

	client, err := NewSSHClient(user, host, port, SetPassword(password))
	if err != nil {
		t.Errorf("NewSSHClient error:%v", err)
	}

	cacheKey := fmt.Sprintf("%s@%s:%d", user, host, port)

	err = clientPool.Put(cacheKey, client)
	if err != nil {
		t.Errorf("Put error:%v", err)
	}

	clientPool.Delete(cacheKey)

	_, ok := clientPool.Get(cacheKey)
	if ok {
		t.Errorf("Test Delete failed")
	}
}

func TestPoolCleanup(t *testing.T) {
	var (
		user     = "root"
		host     = "127.0.0.1"
		port     = 2222
		password = "secret"
	)

	go simpleSSHServer()
	clientPool := NewPool(time.Second*2, time.Second*2, 1000, 3)
	defer clientPool.Close()

	client, err := NewSSHClient(user, host, port, SetPassword(password))
	if err != nil {
		t.Errorf("NewSSHClient error:%v", err)
	}

	cacheKey := fmt.Sprintf("%s@%s:%d", user, host, port)

	err = clientPool.Put(cacheKey, client)
	if err != nil {
		t.Errorf("Put error:%v", err)
	}

	time.Sleep(time.Second * 5)

	_, ok := clientPool.Get(cacheKey)
	if ok {
		t.Errorf("Test Cleanup failed")
	}
}
