/*
Database.go
Yayoi

Created by Cure Gecko on 5/16/15.
Copyright 2015, Cure Gecko. All rights reserved.

Database initialization and structure
*/

package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v1"
	"log"
)

type User struct {
	Id               uint64
	Name             string // `db:"size:50"`
	Email            string // `db:"size:100"`
	Password         []byte // `db:"type:binary size:64"`
	PasswordSalt     []byte // `db:"type:binary size:32"`
	ResetKey         string // `db:"size:30 collation:utf8_bin"`
	SignupKey        string // `db:"size:30 collation:utf8_bin"`
	ApiKey           string // `db:"size:30 collation:utf8_bin"`
	Level            int
	JoinTime         int64
	LastLoginTime    int64
	ResetRequestTime int64
}

type Login struct {
	Ip            string // `db:"size:39"`
	LoginNonce    []byte // `type:binary db:"size:32"`
	LoginAttempts int
	LastAttempt   int64
}

type Authentication struct {
	Token   string // `db:"size:30 collation:utf8_bin"`
	UserID  uint64
	Ip      string // `db:"size:39"`
	Time    int64
	Expires int64
}

func initDb() *gorp.DbMap {
	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@/"+dbName)
	if err != nil {
		log.Fatal(err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"Aira", "UTF8"}}

	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	dbmap.AddTableWithName(Login{}, "login").SetKeys(false, "Ip")
	dbmap.AddTableWithName(Authentication{}, "authentications").SetKeys(false, "Token")

	return dbmap
}
