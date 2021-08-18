# pqssh

Go driver for PostgreSQL over SSH. This driver can connect to postgres on a server via SSH using the local ssh-agent, password, or private-key.

## Why

When using databases, if the database server is in the local network, you can connect directly to it, but the database server is remote and you can only connect to it via ssh. In such cases, it is necessary to use port forwarding and other methods to connect.

However, this requires tunneling with the ssh command beforehand, and it is somewhat time-consuming to operate a remote database in batch mode.

This package provide way to connect the remote postgres via ssh transport.

## Usage

You should call `sql.Register` with your authentication provider information, Hostname, Port, Username, Password, PrivateKey. No need to consider that the connection is over SSH transport.

```go
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/mattn/pqssh"
)

func main() {
	driver := &pqssh.Driver{
		Hostname:   "my-server",
		Port:       22,
		Username:   "sshuser",
		Password:   "sshpassword",
		PrivateKey: "/home/mattn/.ssh/id_rsa",
	}

	sql.Register("postgres+ssh", driver)

	db, err := sql.Open("postgres+ssh", "postgres://my-db-user:my-db-password@127.0.0.1:5432/example?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT id, text FROM example ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d  Name: %s\n", id, name)
	}
}
```

If you don't want to hardcode your credentials, you can use this way:

```go
type config struct {
	pqssh.Driver
	DSN string `json:"dsn"`
}

func main() {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	var cfg config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	sql.Register("postgres+ssh", &cfg)

	// Blah, Blah
}
```

## Requirements

Go

## Installation

```
$ go get github.com/mattn/pqssh
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
