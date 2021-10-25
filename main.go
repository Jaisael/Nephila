package main

import (
	"bufio"
	"database/sql"
	"log"
	"net"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//Connection related data is stored in a sqlite database.
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.Ping()

	//Listen on port 23 - the default telnet port.
	ln, err := net.Listen("tcp", ":23")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print("Couldn't accept connection.'")
			return
		}
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) error {
	for {
		msg, err := bufio.NewReader(c).ReadBytes('\n')
		if err != nil {
			return err
		}
		c.Write(msg)
	}
}
