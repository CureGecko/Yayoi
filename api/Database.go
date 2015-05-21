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

type Setting struct {
	Name  string // `db:"size:255"`
	Value string
}

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
	Ip            string // `db:"size:41"`
	LoginNonce    []byte // `type:binary db:"size:32"`
	LoginAttempts int
	LastAttempt   int64
}

type Authentication struct {
	Token   string // `db:"size:30 collation:utf8_bin"`
	UserID  uint64
	Ip      string // `db:"size:41"`
	Time    int64
	Expires int64
}

type Upload struct {
	UserID             uint64
	MD5                string // `db:"size:32"`
	SHA1               string // `db:"size:40"`
	SHA256             string // `db:"size:64"`
	SHA512             string // `db:"size:128"`
	Extension          string // `db:"size:5"`
	FileSize           int64
	Width              int
	Height             int
	ThumbnailExtension string // `db:"size:5"`
	ThumbnailFileSize  int64
	ThumnailWidth      int
	ThumnailHeight     int
	Rating             string // `db:"size:1"`
	Author             string // `db:"size:50"`
	AuthorURL          string
	SourceURL          string
	Tags               string
	Time               int64
	Submitted          int
}

type Tag struct {
	Id       uint64
	Value    string
	Alias    string
	UseCount int64
}

func initDb() *gorp.DbMap {
	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@/"+dbName)
	if err != nil {
		log.Fatal(err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"Aira", "UTF8"}}

	dbmap.AddTableWithName(Setting{}, "settings").SetKeys(false, "Name")
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	dbmap.AddTableWithName(Login{}, "login").SetKeys(false, "Ip")
	dbmap.AddTableWithName(Authentication{}, "authentications").SetKeys(false, "Token")
	dbmap.AddTableWithName(Upload{}, "uploads").SetKeys(false, "SHA256")
	dbmap.AddTableWithName(Tag{}, "tags").SetKeys(true, "Id")

	return dbmap
}
