package main

import (
	"bufio"
	"database/sql"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var connectionList map[string]*Connection

func main() {
	//Connection related data is stored in a sqlite database.
	dab, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db = dab

	connectionList = make(map[string]*Connection)

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
	loggedin      bool
}

func (c *Connection) Add() {
	connectionList[c.username] = c
}

func (c *Connection) Remove() {
	delete(connectionList, c.username)
}

func handleConnection(c net.Conn) error {
	newConn := &Connection{con: c, state: 1, loggedin: false}
	c.Write([]byte("Welcome."))
	c.Write([]byte("Username:"))
	reader := bufio.NewReader(c)

	//Login dialogue.
	for !newConn.loggedin {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Print(err)
			return err
		}
		msg = strings.TrimSpace(msg)
		switch newConn.state {
		case 1:
			if isValidName(msg) {
				if characterExists(msg) {
					newConn.state = 2
					newConn.username = strings.Title(strings.ToLower(msg))
					c.Write([]byte("Password:"))
				} else {
					newConn.username = strings.Title(strings.ToLower(msg))
					c.Write([]byte("No record exists of a character with this name. Would you like to create a character with this name?"))
					newConn.state = 3
				}
			} else {
				c.Write([]byte("That is not a valid name."))
				c.Write([]byte("Username:"))
			}
		case 2:
			if checkPassword(newConn.username, msg) {
				newConn.loggedin = true
			} else {
				newConn.passwordwrong++
				if newConn.passwordwrong > 2 {
					c.Write([]byte("That is not this character's password. Please try again later."))
					log.Printf("Too many failed login attempts from %v.", newConn.con.RemoteAddr().String())
					c.Close()
				} else {
					c.Write([]byte("That is not the password for this character."))
					c.Write([]byte("What is the password for this character?"))
				}
			}
		case 3:
			log.Print([]byte(msg))
			if strings.ToLower(msg) == "y" || strings.ToLower(msg) == "yes" {
				newConn.state = 4
				c.Write([]byte("Password:"))
			} else {
				newConn.state = 1
				c.Write([]byte("Username:"))
			}
		case 4:
			if msg != "" {
				err := createCharacter(newConn.username, msg)
				if err != nil {
					c.Write([]byte("Difficulty was experienced in creating a new character. Please try again later."))
					c.Close()
				} else {
					newConn.loggedin = true
				}
			}
		case 5:
			newConn.loggedin = true
		default:
			log.Print(msg)
		}
	}

	newConn.Add()
	//Handle commands.
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if msg != "" {
			log.Print(newConn.username, msg)
		}
		c.Write([]byte("tick"))
	}
}

func getPasswordHash(s string) (string, error) {
	rows, err := db.Query("SELECT name, password FROM characters")
	if err != nil {
		log.Print(err)
		return "", err
	}
	var name string
	var password string
	for rows.Next() {
		err = rows.Scan(&name, &password)
		if err != nil {
			log.Print(err)
			return "", err
		}
		if s == name {
			return password, nil
		}
	}
	return "", nil
}

func checkPassword(nm, pwd string) bool {
	hashP, err := getPasswordHash(nm)
	if err != nil {
		log.Println(err)
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashP), []byte(pwd))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func isValidName(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func characterExists(s string) bool {
	rows, err := db.Query("SELECT name FROM characters")
	if err != nil {
		log.Print(err)
		return false
	}
	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			log.Print(err)
			return false
		}
		if s == name {
			return true
		}
	}
	return false
}

func createCharacter(name, pwd string) error {
	//Passwords are hashed and salted using the bcrypt package before being stored.
	hashP, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		log.Print(err)
		return err
	}

	stmt, err := db.Prepare("INSERT INTO characters (name, password, banned, lastonline) VALUES (?,?,?,?)")
	if err != nil {
		log.Print(err)
		return err
	}
	res, err := stmt.Exec(name, string(hashP), 0, time.Now())
	if err != nil {
		log.Print(err)
		return err
	}
	log.Print(res)
	return nil
}
