package model

import (
	"database/sql"
	"fmt"
	"encoding/json"
)

// Info contains the database configurations
type DBInfo struct {
	// Database type
	Type string
}

// Configuration contains the application settings
type Configuration struct {
	Database DBInfo         `json:"Database"`
	Server   Server   `json:"Server"`
}

// ParseJSON unmarshals bytes to structs
func (c *Configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

// Server stores the hostname and port number
type Server struct {
	UseHTTP   bool   `json:"UseHTTP"`   // Listen on HTTP
	UseHTTPS  bool   `json:"UseHTTPS"`  // Listen on HTTPS
	HTTPPort  string    `json:"HTTPPort"`  // HTTP port
	HTTPSPort string    `json:"HTTPSPort"` // HTTPS port
	CertFile  string `json:"CertFile"`  // HTTPS certificate
	KeyFile   string `json:"KeyFile"`   // HTTPS private key
}


type Realm string

type Right int

const (
	SYSADMIN = 1 + iota
	ADMIN
	PROVISIONER
	READONLY
)

var rights = [...]string{
	"Sys Admin",
	"Admin",
	"Provisioner",
	"Read Only",

}

func (right Right) String() string {
	return rights[right-1]
}

type User struct {
	Username string `json:"uname"`
	Password string `json:"password"`
	Realm    Realm  `json:"realm"`
	Right    Right  `json:"uright"`
}

func GetUsers(db *sql.DB) ([]User, error) {
	// TODO : Add LIMIT and OFFSET support
	rows, err := db.Query("SELECT uname, password, realm, uright FROM users")

	if err != nil {
		fmt.Printf("Could not get all User's for (error : %s)", err.Error())
		return nil, err
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Username, &u.Password, &u.Realm, &u.Right); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil

}

func (u *User) GetUser(db *sql.DB) error {
	err := db.QueryRow("SELECT uname, password, realm, uright FROM users WHERE uname=$1", u.Username).Scan(
		&u.Username, &u.Password, &u.Realm, &u.Right)
	if err != nil {
		fmt.Printf("ERROR : %s\n", err.Error())
	}
	return err
}

func (u *User) UpdateUser(db *sql.DB) error {
	err := db.QueryRow("UPDATE users SET uname=$1, password=$2, realm=$3, right=$4"+
		"WHERE uname=$1 RETURNING uname, password, realm",
		u.Username, u.Password, u.Realm).Scan(&u.Username, &u.Password, &u.Realm, &u.Right)
	return err
}

func (u *User) DeleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE uname=$1", u.Username)
	if err != nil {
		fmt.Printf("ERROR : %s\n", err.Error())
	}
	return err
}

func (u *User) CreateUser(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO users(uname, password, realm) VALUES($1, $2, $3, $4) "+
			"RETURNING uname", u.Username, u.Password, u.Realm, u.Right).Scan(&u.Username)
	if err != nil {
		fmt.Printf("Could not create User")
		return err
	}
	return nil
}