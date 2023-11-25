# ssh-client-pool
ssh-client-pool is a simple ssh client pool for golang.

### Installation
`go get github.com/WeiZhixiong/ssh-client-pool`

### Usage
```go
package main

import (
    "fmt"
    "time"
	pool "github.com/WeiZhixiong/ssh-client-pool"
)

func main() {
	var (
		user     = "root"
		host     = "127.0.0.1"
		port     = 2222
		password = "secret"

	)
	sshClientPool := pool.NewPool(time.Minute*10, time.Minute*10, 500, 1)

	client, err := pool.NewSSHClient(user, host, port, pool.SetPassword(password))
	if err != nil {
		fmt.Printf("NewSSHClient error:%v", err)
		return
	}
	// client do something
	
	cacheKey := fmt.Sprintf("%s@%s:%d", user, host, port)
	err = sshClientPool.Put(cacheKey, client)
	if err != nil {
		fmt.Printf("Put error:%v\n", err)
		_ = client.Close()
		return
	}

	client, ok := sshClientPool.Get(cacheKey)
	if ok {
		fmt.Printf("Get error:%v\n", err)
	}

	// or use GetWithNew
	client, err = sshClientPool.GetWithNew(cacheKey, user, host, port, pool.SetPassword(password))
	if err != nil {
		fmt.Printf("GetWithNew error:%v\n", err)
		return
	}
	
	// client do something
	
	// put back to pool
	err = sshClientPool.Put(cacheKey, client)
	if err != nil {
		fmt.Printf("Put error:%v\n", err)
		_ = client.Close()
		return
	}
	
	// delete from pool
	sshClientPool.Delete(cacheKey)
	
	// close pool
	sshClientPool.Close()
}
```