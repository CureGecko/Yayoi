/*
Auth.go
Yayoi

Created by Cure Gecko on 5/15/15.
Copyright 2015, Cure Gecko. All rights reserved.

Manages user authentication.
*/
package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

type Auth struct {
	Server        *Iori
	Writer        http.ResponseWriter
	Request       *http.Request
	Authenticated bool
	ID            int64
	Token         string
	Name          string
	Level         uint
}

func (a *Auth) Validate() {
	a.Authenticated = false
	now := time.Now().Unix()
	var token string
	for _, cookie := range a.Request.Cookies() {
		if cookie.Name == "IoriAuth" {
			token = cookie.Value
		}
	}

	var id int64
	var ip string
	var expires int64
	err := a.Server.DB.QueryRow("SELECT `userID`,`ip`,`expires` FROM authentications WHERE `token`=? AND `expires`>=?", token, now).Scan(&id, &ip, &expires)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Auth:", err)
		return
	} else if err == sql.ErrNoRows {
		return
	}
	var name string
	var level uint
	err = a.Server.DB.QueryRow("SELECT `name`,`level` FROM users WHERE `id`=?", id).Scan(&name, &level)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Auth:", err)
		return
	} else if err == sql.ErrNoRows {
		return
	}
	a.Authenticated = true
	a.ID = id
	a.Token = token
	a.Name = name
	a.Level = level

	if expires <= now+(60*10) /* 10 minutes */ {
		newToken := randomString(30)
		_, err := a.Server.DB.Exec("UPDATE authentications SET `token`=?, `expires`=? WHERE `token`=?", newToken, now+(60*60*24) /* 1 day */, token)
		if err != nil {
			log.Println("Auth:", err)
			return
		}
		a.Token = newToken

		IoriAuth := new(http.Cookie)
		IoriAuth.Name = "IoriAuth"
		IoriAuth.Value = newToken
		IoriAuth.Path = SitePath
		IoriAuth.Expires = time.Now().Add(time.Hour * 24 /* 1 day */)
		IoriAuth.HttpOnly = true
		http.SetCookie(a.Writer, IoriAuth)
	}
}
