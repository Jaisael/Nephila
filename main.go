package main

import (
	"bufio"
	"database/sql"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//Connection related data is stored in a sqlite database.
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

type Connection struct {
	con           net.Conn
	username      string
	passwordwrong int
	state         int
}

func handleConnection(c net.Conn) error {
	newConn := &Connection{con: c, state: 1}
	c.Write([]byte("Welcome."))
	c.Write([]byte("Username:"))
	reader := bufio.NewReader(c)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		switch newConn.state {
		case 1:
			if msg != "" {
				if characterExists(msg) {
					newConn.state = 2
					newConn.username = strings.Title(strings.ToLower(msg))
					c.Write([]byte("Password:"))
				} else {
					c.Write([]byte("No record exists of a character with this name. Would you like to create a character with this name?"))
					newConn.state = 3
				}

			} else {
				c.Write([]byte("Username:"))
			}
		case 2:
			if checkPassword(msg, "") {
				newConn.state = 4
			} else {
				newConn.passwordwrong++
				if newConn.passwordwrong > 2 {
					c.Write([]byte("That is not this character's password. Please try again later."))
					c.Close()
				} else {
					c.Write([]byte("That is not the password for this character."))
					c.Write([]byte("What is the password for this character?"))
				}
			}
		case 3:

		default:
			log.Print(msg)
		}
	}
}

func hashPassword(p string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func checkPassword(p, h string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(p))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func characterExists(s string) bool {
	return true
}
