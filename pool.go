package pool

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

type Item struct {
	Object     *ssh.Client
	Expiration int64
}

// Expired Returns true if the item has expired.
func (item Item) Expired() bool {
	return time.Now().Unix() > item.Expiration
}

type Pool struct {
	items             map[string]chan Item
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	maxConn           int
	eachKeyMaxConn    int
	currentConn       int
	cleanerStop       chan struct{}
	mu                sync.Mutex
}

// NewPool creates a new pool.
// defaultExpiration: The default expiration time.
// cleanupInterval: How often the pool is checked for expired items.
// maxConn: The maximum number of connections that can be held by the pool.
// eachKeyMaxConn: The maximum number of connections that can be held by the pool for each key.
func NewPool(defaultExpiration, cleanupInterval time.Duration, maxConn, eachKeyMaxConn int) *Pool {
	if maxConn == 0 {
		maxConn = 1000
	}
	if eachKeyMaxConn == 0 {
		eachKeyMaxConn = 3
	}
	if defaultExpiration <= 0 {
		defaultExpiration = 10 * time.Minute
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 10 * time.Minute
	}
	items := make(map[string]chan Item, 8)
	pool := &Pool{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		maxConn:           maxConn,
		eachKeyMaxConn:    eachKeyMaxConn,
		cleanerStop:       make(chan struct{}),
	}

	go pool.cleaner()

	return pool
}

func (p *Pool) Len() int {
	return p.currentConn
}

// Put adds an ssh Client to the pool
// key: suggested key is user + host + port
func (p *Pool) Put(key string, c *ssh.Client) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.currentConn >= p.maxConn {
		return errors.New("pool is full")
	}
	if _, ok := p.items[key]; !ok {
		p.items[key] = make(chan Item, p.eachKeyMaxConn)
	}
	if len(p.items[key]) >= p.eachKeyMaxConn {
		return errors.New("key is full")
	}
	p.items[key] <- Item{
		Object:     c,
		Expiration: time.Now().Add(p.defaultExpiration).Unix(),
	}
	p.currentConn++
	return nil
}

func (p *Pool) get(key string) (*Item, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.items[key]; !ok {
		return nil, false
	}

	itemNum := len(p.items[key])
	for i := 0; i < itemNum; i++ {
		item := <-p.items[key]
		p.currentConn--
		if len(p.items[key]) == 0 {
			delete(p.items, key)
		}
		if item.Expired() {
			_ = item.Object.Close()
			continue
		}
		return &item, true
	}

	return nil, false
}

// Get return ssh client from the pool by the key, and the ssh client will be removed from the pool.
func (p *Pool) Get(key string) (*ssh.Client, bool) {
	for {
		item, ok := p.get(key)
		if !ok {
			return nil, false
		}
		// if the ssh client put into the pool less than 60 seconds, skip the keepalive test.
		if time.Now().Add(p.defaultExpiration).Unix()-item.Expiration < 60 {
			return item.Object, true
		}
		err := KeepAlive(item.Object)
		if err != nil {
			_ = item.Object.Close()
			continue
		}
		return item.Object, true
	}
}

func (p *Pool) Delete(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.items[key]; !ok {
		return
	}
	itemNum := len(p.items[key])
	for i := 0; i < itemNum; i++ {
		item := <-p.items[key]
		p.currentConn--
		_ = item.Object.Close()
		continue
	}
	return
}

// GetWithNew return ssh client from the pool, if there is no client in the pool, it will create a new one.
func (p *Pool) GetWithNew(key, user, host string, port int, opts ...ClientCfgOption) (c *ssh.Client, err error) {
	c, ok := p.Get(key)
	if ok {
		return c, nil
	}
	c, err = NewSSHClient(user, host, port, opts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (p *Pool) clean() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for key, itemCh := range p.items {
		for i := 0; i < len(itemCh); i++ {
			item := <-itemCh
			if item.Expired() {
				_ = item.Object.Close()
				continue
			}
			itemCh <- item
		}
		if len(itemCh) == 0 {
			delete(p.items, key)
		}
	}
}

func (p *Pool) cleaner() {
	ticker := time.NewTicker(p.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.clean()
		case <-p.cleanerStop:
			return
		}
	}
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.cleanerStop)

	for _, itemCh := range p.items {
		for i := 0; i < len(itemCh); i++ {
			item := <-itemCh
			_ = item.Object.Close()
		}
	}
	p.items = nil
}
