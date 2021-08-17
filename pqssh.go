// pqssh is a Go driver for PostgreSQL over SSH. This driver can connect
// to postgres on a server via SSH using the local ssh-agent, password, or
// private-key.

package pqssh

import (
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Driver is authentication provider information.
type Driver struct {
	Hostname   string `json:"hostname"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	client     *ssh.Client
	agent      agent.Agent
}

// Open opens connection to the server.
func (d *Driver) Open(s string) (driver.Conn, error) {
	sshConfig := &ssh.ClientConfig{
		User:            d.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		defer conn.Close()
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
	}

	if d.PrivateKey != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return getSigners(d.PrivateKey, d.Password)
		}))
	} else if d.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PasswordCallback(func() (string, error) {
			return d.Password, nil
		}))
	}

	sshcon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", d.Hostname, d.Port), sshConfig)
	if err != nil {
		return nil, err
	}
	d.client = sshcon

	return pq.DialOpen(d, s)
}

// Dial make socket connection via SSH.
func (d *Driver) Dial(network, address string) (net.Conn, error) {
	return d.client.Dial(network, address)
}

// DialTimeout make socket connection via SSH.
func (d *Driver) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return d.client.Dial(network, address)
}

func getSigners(keyfile string, password string) ([]ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	if password != "" {
		k, err := ssh.ParsePrivateKeyWithPassphrase(buf, []byte(password))
		if err != nil {
			return nil, err
		}
		return []ssh.Signer{k}, nil
	}

	k, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return []ssh.Signer{k}, nil
}
