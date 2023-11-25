package pool

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

type ClientCfgOption func(c *ssh.ClientConfig)

func SetPassword(password string) ClientCfgOption {
	return func(c *ssh.ClientConfig) {
		if c.Auth == nil {
			c.Auth = make([]ssh.AuthMethod, 0, 2)
		}
		c.Auth = append(c.Auth, ssh.Password(password))
	}
}

func SetPrivateKey(privateKey, passPhrase string) ClientCfgOption {
	return func(c *ssh.ClientConfig) {
		var (
			signer ssh.Signer
			err    error
		)

		if c.Auth == nil {
			c.Auth = make([]ssh.AuthMethod, 0, 2)
		}

		if passPhrase != "" {
			if signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passPhrase)); err == nil {
				c.Auth = append(c.Auth, ssh.PublicKeys(signer))
				return
			}
		}

		if signer, err = ssh.ParsePrivateKey([]byte(privateKey)); err == nil {
			c.Auth = append(c.Auth, ssh.PublicKeys(signer))
			return
		}
	}
}

func SetTimeout(seconds int) ClientCfgOption {
	return func(c *ssh.ClientConfig) {
		c.Timeout = time.Duration(seconds) * time.Second
	}
}

func NewSSHClient(user, host string, port int, opts ...ClientCfgOption) (*ssh.Client, error) {
	var (
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	clientConfig = &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for _, opt := range opts {
		opt(clientConfig)
	}

	addr = fmt.Sprintf("%s:%d", host, port)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return sshClient, nil
}

func KeepAlive(c *ssh.Client) error {
	_, _, err := c.SendRequest("keepalive@openssh.com", false, nil)
	return err
}
