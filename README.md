# pqssh

Go driver for PostgreSQL over SSH. This driver can connect to postgres on a server via SSH using the local ssh-agent, password, or private-key.

## Usage

```go
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/mattn/pqssh"
)

func main() {
	driver := &pqssh.PqSshDriver{
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
