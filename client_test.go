package pool

import (
	"github.com/gliderlabs/ssh"
	"testing"
)

func simpleSSHServer() {
	pubKey := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCxIDnavN3b+9dr5RkcV5ftoU+19dLY9XJqBp84ZL+7dmt/orNTXwVAnXVVptG9BlOSHzixr42z5CItkLiPeTM15JeKSv/SDv8pfrj8ZWZqX4RtsSFyMMIkj6dUAiNzUQrTK9CLD+4rVtAfvX2++5gcSR+r2+0YyTq23cYfVevJNJuEZ3Lct1uL0BARyrCd3HHnGbOIiRKT4slKaKdGd/g/1HxwJbzZ3CpywCBSLjB2Sl2NLTezt+mdQR80HMEHTB0oNZiSOEzwOvkzZWvtfFpY5xwZ6shBWls+/9/yUz8jDCwva+xfSwZJjsQXIvrmj8bTYqI2MVii3UQSn5bhvU+SuRAQwp5q/TCR6YQPACqzor77lesGLYfZJTtX6dCy+5qy9i6ArC+/GbCxg+HZP/CRcKRu3ubUwL/h/yCf8fZ1X6/K6iREEqAK9t9BUPTNxdPju3LcAih0efPT7vJTDRpFJr5JuLWzI0p0XLCOVTGzzETqwCixFaEC/YktDPpynuc= test3@127.0.0.1`
	passAuthOption := ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
		if ctx.User() == "root" && pass == "secret" {
			return true
		}
		if ctx.User() == "test" && pass == "secret1" {
			return true
		}
		return false
	})
	publicKeyAuthOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		allowed, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(pubKey))
		return ctx.User() == "test3" && ssh.KeysEqual(key, allowed)
	})
	ssh.ListenAndServe("127.0.0.1:2222", nil, passAuthOption, publicKeyAuthOption)
}

func TestNewSSHClientWithPassword(t *testing.T) {
	go simpleSSHServer()
	_, err := NewSSHClient("root", "127.0.0.1", 2222, SetPassword("secret"))
	if err != nil {
		t.Errorf("NewSSHClient error:%v", err)
	}
}

func TestNewSSHClientWithPrivateKey(t *testing.T) {
	privateKey := `
-----BEGIN OPENSSH PRIVATE KEY-----
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
-----END OPENSSH PRIVATE KEY-----`

	go simpleSSHServer()
	_, err := NewSSHClient("test3", "127.0.0.1", 2222, SetPrivateKey(privateKey, ""))
	if err != nil {
		t.Errorf("NewSSHClient error:%v", err)
	}
}

func TestKeepAlive(t *testing.T) {
	var (
		user     = "root"
		host     = "127.0.0.1"
		port     = 2222
		password = "secret"
	)

	go simpleSSHServer()
	client, err := NewSSHClient(user, host, port, SetPassword(password))
	if err != nil {
		t.Errorf("NewSSHClient error:%v", err)
	}

	err = KeepAlive(client)
	if err != nil {
		t.Errorf("KeepAlive error:%v", err)
	}
}
