package pqssh

import (
	"crypto/x509"
	"database/sql/driver"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type PqSshDriver struct {
	ViaHostname   string
	ViaPort       int
	ViaUsername   string
	ViaPassword   string
	ViaPrivateKey string
	client        *ssh.Client
	agent         agent.Agent
}

func (d *PqSshDriver) Open(s string) (driver.Conn, error) {
	sshConfig := &ssh.ClientConfig{
		User:            d.ViaUsername,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		defer conn.Close()
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
	}

	if d.ViaPrivateKey != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return getSigners(d.ViaPrivateKey, d.ViaPassword)
		}))
	} else if d.ViaPassword != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PasswordCallback(func() (string, error) {
			return d.ViaPassword, nil
		}))
	}

	sshcon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", d.ViaHostname, d.ViaPort), sshConfig)
	if err != nil {
		return nil, err
	}
	d.client = sshcon

	return pq.DialOpen(d, s)
}

func (d *PqSshDriver) Dial(network, address string) (net.Conn, error) {
	return d.client.Dial(network, address)
}

func (d *PqSshDriver) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return d.client.Dial(network, address)
}

func getSigners(keyfile string, password string) ([]ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(buf)
	if x509.IsEncryptedPEMBlock(b) {
		buf, err = x509.DecryptPEMBlock(b, []byte(password))
		if err != nil {
			return nil, err
		}
		pk, err := x509.ParsePKCS1PrivateKey(buf)
		if err != nil {
			return nil, err
		}
		k, err := ssh.NewSignerFromKey(pk)
		if err != nil {
			return nil, err
		}
		return []ssh.Signer{k}, nil
	}
	k, err := ssh.ParsePrivateKey(buf)
	if err == nil {
		return []ssh.Signer{k}, nil
	}
	return nil, err
}
